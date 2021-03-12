// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE. }}}

package rtltcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/rtl"
	"hz.tools/sdr/rtl/e4k"
	"hz.tools/sdr/stream"
)

var (
	// ErrSDRNotFound will be returned if no SDR can be acquired.
	ErrSDRNotFound error = fmt.Errorf("rtltcp: SDR Not Found")
)

// ServerHandler will return an SDR to be used by the incoming
// connection.
type ServerHandler func(context.Context) (sdr.Receiver, error)

// CommandHandler will handle incoming requests and process them
type CommandHandler func(context.Context, sdr.Receiver, Request) error

// Server encapsulates internal state to listen for and handle incoming
// requests from the client.
type Server struct {
	// (Optional) TCP address to listen on.
	Addr string

	// Handler will be called when a new request comes in, and be used to create
	// the sdr.Receiver to be used by the Server runtime, and stream IQ samples
	// to the remote end.
	//
	// TODO(paultag): rename Handler
	Handler ServerHandler

	// CommandHandler will handle incoming requests and process them. If nil,
	// the default handler will be used.
	CommandHandler CommandHandler

	// (Optional) If Handler is not set, this value is used to tell the
	// DefaultCommandHandler what gain stage to control.
	GainStageName string

	// (Optional) If Handler is not set, this value is used to tell the
	// DefaultCommandHandler what IF gain stage to control if the device's
	// tuner is an e4k.
	IFGainStageName string

	// ConnContext will create a context based on the provided net.Conn
	ConnContext func(ctx context.Context, c net.Conn) context.Context
}

// NewDefaultCommandHandler will create the default rtltcp CommandHandler
// connected to the provided GainStage and IF GainStage.
func NewDefaultCommandHandler(defaultGainStageName, defaultIFGainStageName string) CommandHandler {
	gainState := e4k.Stages{}

	return func(ctx context.Context, dev sdr.Receiver, request Request) error {
		arg := request.Argument
		switch request.Command {
		case CommandSetFreq:
			log.Printf("Setting freq to %s\n", rf.Hz(arg))
			return dev.SetCenterFrequency(rf.Hz(arg))
		case CommandSetSampleRate:
			log.Printf("Setting SampleRate to %d\n", arg)
			return dev.SetSampleRate(uint(arg))
		case CommandSetGainMode:
			log.Printf("Setting Gain mode %d\n", arg)
			return dev.SetAutomaticGain(arg == 0)
		case CommandSetGain:
			gain := 0.1 * float32(arg)
			log.Printf("setting gain %f %d\n", gain, arg)
			return sdr.SetGainStages(dev, map[string]float32{
				defaultGainStageName: gain,
			})
		case CommandSetIFGain:
			if defaultIFGainStageName == "" {
				log.Printf("no IF gain stage, not adjusting IF gain\n")
				return nil
			}
			gain := int16(arg & 0xFFFF)
			stage := (arg >> 16) - 1
			log.Printf("gain stage %d set to %d\n", stage, gain)
			if stage > 6 {
				log.Printf("Malformed IF Gain request: stage=%d, gain=%d\n", stage, gain)
				return nil
			}
			gainState[stage] = int(gain)
			log.Printf("Virtual IF Gain state: %d, total gain: %f\n", gainState, gainState.GetGain())
			return sdr.SetGainStages(dev, map[string]float32{
				defaultIFGainStageName: gainState.GetGain(),
			})
		case CommandSetBiasTee:
			// TODO(paultag): This one may be worth implementing.
			return nil
		case CommandSetAGCMode, CommandSetDirectSampling, CommandSetOffsetTuning:
			// Ignore!
			return nil
		default:
			log.Printf("Unsupported command: %x (%s)\n", request.Command, request.Command)
		}

		return nil
	}
}

// Tunerable is an interface that allows the Sdr to specify what kind of
// RTL-SDR Tuner is being used.
type Tunerable interface {
	Tuner() rtl.Tuner
}

func (s Server) serveConn(ctx context.Context, conn net.Conn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer conn.Close()
	defer cancel()

	if s.ConnContext != nil {
		ctx = s.ConnContext(ctx, conn)
	}

	dev, err := s.Handler(ctx)
	if err != nil {
		log.Printf("Error accepting new connection - closing connection")
		log.Println(err)
		return err
	}
	defer dev.Close()

	tuner := rtl.TunerUnknown
	tunerable, ok := dev.(Tunerable)
	if ok {
		tuner = tunerable.Tuner()
		log.Printf("Tuner detected as %s\n", tuner)
	}

	// TunerInfo
	if err := binary.Write(conn, binary.BigEndian, &DongleInfo{
		Magic:     [4]byte{'R', 'T', 'L', '0'},
		TunerType: uint32(tuner),
	}); err != nil {
		log.Printf("Error writing DongleInfo\n")
		log.Println(err)
		return err
	}

	handler := s.CommandHandler
	if handler == nil {
		handler = NewDefaultCommandHandler(
			s.GainStageName,
			s.IFGainStageName,
		)
	}

	reader, err := dev.StartRx()
	if err != nil {
		log.Printf("Error starting SDR Receiver\n")
		log.Println(err)
		return err
	}
	defer reader.Close()

	u8Reader, err := stream.ConvertReader(reader, sdr.SampleFormatU8)
	if err != nil {
		log.Printf("Error creating conversion reader\n")
		log.Println(err)
		cancel()
		return err
	}

	writer := sdr.ByteWriter(conn, binary.LittleEndian, 0, sdr.SampleFormatU8)

	go func() {
		defer cancel()
		req := Request{}
		for {
			if ctx.Err() != nil {
				log.Printf("Context error, aborting goroutine\n")
				log.Println(err)
				return
			}
			if err := binary.Read(conn, binary.BigEndian, &req); err != nil {
				log.Printf("Error reading command; discarding\n")
				log.Println(err)
				if err == io.EOF {
					return
				}
				continue
			}
			log.Printf("%#v\n", req)
			if err := handler(ctx, dev, req); err != nil {
				log.Printf("Error processing command; discarding\n")
				log.Printf("%#v\n", err)
				continue
			}
		}
	}()

	_, err = sdr.Copy(writer, u8Reader)
	if err != nil {
		log.Printf("Error copying samples\n")
		log.Println(err)
		return err
	}

	return nil
}

// Serve will accept connections from the provided listener, and serve
// client requests.
func (s Server) Serve(listener net.Listener) error {
	ctx := context.TODO()
	// TODO: Have this configurable in the Server struct, and augment this
	// with peer info.

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.serveConn(ctx, conn)
	}
}

// ListenAndServe will listen for incoming requests and return them as required.
func (s Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// vim: foldmethod=marker
