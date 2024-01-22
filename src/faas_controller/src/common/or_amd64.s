#include "textflag.h"

// func Or(addr *uint32, v uint32)
TEXT ·Or(SB), NOSPLIT, $0-12
	MOVQ	ptr+0(FP), AX
	MOVL	val+8(FP), BX
	LOCK
	ORL	BX, (AX)
	RET
