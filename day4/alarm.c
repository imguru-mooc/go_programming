#if 1
#include <unistd.h>
#include <stdio.h>
#include <signal.h>

void my_sig(int signo)
{
	printf("my_sig(%d)\n", signo);
	alarm(2);
}

int main()
{
	signal(SIGALRM, my_sig);

	alarm(2);
	while(1)
	{
		sleep(1);
		printf(".\n");
	}
	return 0;
}
#endif
#if 0
#include <unistd.h>
#include <stdio.h>
#include <signal.h>

void my_sig(int signo)
{
	printf("my_sig(%d)\n", signo);
}

int main()
{
	char buff[1024];
	int ret;
	signal(SIGALRM, my_sig);

	while(1)
	{
		alarm(3);
		ret = read(0, buff, sizeof buff);
		buff[ret-1] = 0;
		printf("keyboard=[%s]\n", buff);
		alarm(0);
	}

	return 0;
}
#endif
