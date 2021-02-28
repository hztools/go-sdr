# AMD64 Registers

Using exotic registers like R8->R13 can cause weird issues with perf. It's best
to stick with the most common registers below for perf critical segments.

|Reg | Description                             |
|----|-----------------------------------------|
| AX | GP                                      |
| BX | GP, Often a const value used throughout |
| CX | GP, Function param or loop counter      |
| DX | GP, Function param, storing temp vars   |
| SI | GP, Often a pointer.                    |
| DI | GP, Often a pointer                     |
