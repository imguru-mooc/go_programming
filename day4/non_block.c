#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdlib.h>
#include <errno.h>
#include <sys/select.h>

int main()
{
	int ret;
	int fd;
	char buff[1024];


	fd = open("myfifo", O_RDWR);  // mkfifo myfifo

	fcntl(  0, F_SETFL, O_NONBLOCK);
	fcntl( fd, F_SETFL, O_NONBLOCK);

	while(1)
	{
		ret = read(0, buff, sizeof buff );
		if( ret > 0 )
		{
			buff[ret-1] = 0;
			printf("keyboard=[%s]\n", buff );
		} else if(ret < 0 )
		{
			if( errno != EAGAIN )
				exit(0);
		}

		ret = read(fd, buff, sizeof buff );
		if( ret > 0 )
		{
			buff[ret-1] = 0;
			printf("myfio   =[%s]\n", buff );
		} 
		else if(ret < 0 )
		{
			if( errno != EAGAIN )
				exit(0);
		}
	}
	return 0;
}















