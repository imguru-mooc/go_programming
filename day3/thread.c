#include <pthread.h>
#include <stdio.h>

int global=6;
void * foo(void *data)
{
    printf("foo(), global=%d\n", ++global);
	return 0; // pthread_exit(0);
}

int main()
{
    pthread_t thread;

    pthread_create(&thread, 0, foo, 0);

	pthread_join(thread, 0);
    printf("main(), global=%d\n", global);
}

