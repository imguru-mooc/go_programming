#include <pthread.h>
#include <stdio.h>

int counter;

pthread_mutex_t mu = PTHREAD_MUTEX_INITIALIZER;

void* increment(void* data) {
	int i;
	for(i=0; i<1000000; i++)
	{
		pthread_mutex_lock(&mu);
		counter++;
		pthread_mutex_unlock(&mu);
	}
}

int main()
{
    pthread_t thread[10];
	int i;

	for(i=0; i<10; i++)
    	pthread_create(&thread[i], 0, increment, 0);

	for(i=0; i<10; i++)
		pthread_join(thread[i], 0);

    printf("main(), counter=%d\n", counter);
}

