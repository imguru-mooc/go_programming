#include <stdio.h>
#include <stdlib.h>
#include <ucontext.h>  // 유저 모드 컨텍스트 스위칭을 위한 핵심 라이브러리

// 유저 스레드(고루틴 역할)의 CPU 상태(레지스터, 스택 등)를 저장할 구조체
ucontext_t ctx_main, ctx_g1, ctx_g2;

// 가상의 고루틴 G1 함수
void func_g1() {
    printf("[G1] 1단계 작업을 시작합니다.\n");
    
    // Go의 채널 대기나 자원 양보(runtime.Gosched())와 같은 역할
    printf("[G1] 내부 요인 발생! 작업을 양보하고 G2에게 순서를 넘깁니다.\n");
    swapcontext(&ctx_g1, &ctx_g2); // 현재 CPU 상태를 ctx_g1에 저장하고, ctx_g2 상태를 로드!
    
    // G2가 양보하면 이 자리로 돌아옵니다.
    printf("[G1] 다시 돌아왔습니다! 2단계 마무리 작업을 합니다.\n");
    
    // 메인 함수(스케줄러)로 돌아갑니다.
    swapcontext(&ctx_g1, &ctx_main);
}

// 가상의 고루틴 G2 함수
void func_g2() {
    printf("[G2] G1이 양보해줘서 실행되었습니다.\n");
    printf("[G2] 내 할 일을 다 했으니 다시 G1을 깨웁니다.\n");
    
    swapcontext(&ctx_g2, &ctx_g1); // 현재 CPU 상태를 ctx_g2에 저장하고, ctx_g1 상태를 로드!
}

int main() {
    // 1. 각 유저 스레드가 사용할 독립된 스택 메모리를 '유저 영역'에 할당합니다.
    char stack_g1[1024 * 64];
    char stack_g2[1024 * 64];

    // 2. G1의 가상 스레드 환경 설정
    getcontext(&ctx_g1);                         // 기본 CPU 세팅 복사
    ctx_g1.uc_stack.ss_sp = stack_g1;            // 독립된 스택 매핑
    ctx_g1.uc_stack.ss_size = sizeof(stack_g1);  // 스택 크기 지정
    ctx_g1.uc_link = &ctx_main;                  // 함수가 끝나면 돌아올 곳 지정
    makecontext(&ctx_g1, func_g1, 0);            // PC(프로그램 카운터)를 func_g1 주소로 설정

    // 3. G2의 가상 스레드 환경 설정
    getcontext(&ctx_g2);
    ctx_g2.uc_stack.ss_sp = stack_g2;
    ctx_g2.uc_stack.ss_size = sizeof(stack_g2);
    ctx_g2.uc_link = &ctx_main;
    makecontext(&ctx_g2, func_g2, 0);

    printf("[Main] 유저 스레드 제어 환경 설정 완료.\n");
    printf("[Main] 스케줄러 구동: G1을 먼저 실행합니다.\n");
    printf("--------------------------------------------------\n");

    // 4. 메인 스레드의 상태를 ctx_main에 저장하고, ctx_g1으로 점프합니다.
    swapcontext(&ctx_main, &ctx_g1);

    printf("--------------------------------------------------\n");
    printf("[Main] 모든 유저 스레드가 종료되어 메인 스케줄러로 복귀했습니다.\n");
    return 0;
}
