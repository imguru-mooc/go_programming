#include <stdio.h>
#include <setjmp.h>

jmp_buf env;

void deep_function(void) {
    longjmp(env, 1);  // env로 점프
	//printf("deep_function()\n");
}

int main(void) {
    if (setjmp(env) == 0) {
        deep_function();
    } else {
        printf("점프 후 복귀\n");
    }
}
