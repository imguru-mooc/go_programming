#if 0
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <stdlib.h>
#include <wait.h>
#include <signal.h>

void my_sig(int signo)
{
	int status;
	while(wait(&status)>0)
		printf("status=%d\n", (status>>8)&0xff );
}

int main()
{
	int i,j;
	int pid;

	signal( SIGCHLD, my_sig );

	for(i=0; i<10; i++)
	{
		if (fork() == 0 )
		{
			for(j=0; j<3; j++)
			{
				sleep(1);
				printf("child\n");
			}
			exit(i);
		}
	}

	while(1)
	{
		sleep(1);
		printf("parent\n");
	}
}

#endif
#if 0
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <stdlib.h>
#include <wait.h>
#include <signal.h>

void my_sig(int signo)
{
	int status;
	wait(&status);
	printf("status=%d\n", (status>>8)&0xff );
}

int main()
{
	int i;
	int pid;
			{
			}
		}



	signal( SIGCHLD, my_sig );

	pid = fork();

	if( pid == 0)
	{
		for(i=0; i<3; i++)
		{
			sleep(1);
			printf("child\n");
		}
		exit(7);
	}

	while(1)
	{
		sleep(1);
		printf("parent\n");
	}
}

#endif

#if 0
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <stdlib.h>

int main()
{
	int i;
	int pid;
	int status;
	pid = fork();

	if( pid == 0)
	{
		for(i=0; i<3; i++)
		{
			sleep(1);
			printf("child\n");
		}
		exit(7);
	}

	while(1)
	{
		wait(&status);
		printf("status=%d\n", (status>>8)&0xff );
		sleep(1);
		printf("parent\n");
	}
}

#endif
#if 0
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>
#include <stdlib.h>

int main()
{
	int i;
	int pid;
	pid = fork();

	if( pid == 0)
	{
		for(i=0; i<3; i++)
		{
			sleep(1);
			printf("child\n");
		}
		exit(7);
	}

	while(1)
	{
		sleep(1);
		printf("parent\n");
	}
}

#endif


#if 0
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>

int main()
{
	int pid;
	pid = fork();
	if( pid > 0)
		printf("parent\n");
	else if( pid == 0)
		printf("child\n");
}
#endif


#if 0
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>

int main()
{
	fork();
	printf("after\n");
}
#endif
