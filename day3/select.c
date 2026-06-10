#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/select.h>

int main()
{
	int ret;
	int fd;
	char buff[1024];
	fd_set readfds;

	FD_ZERO(&readfds);

	fd = open("myfifo", O_RDWR);  // mkfifo myfifo

	while(1)
	{
		FD_SET( 0, &readfds);
		FD_SET( fd, &readfds);
		select( fd+1, &readfds,0,0,0 );

		if( FD_ISSET(0, &readfds) )
		{
			ret = read(0, buff, sizeof buff );
			buff[ret-1] = 0;
			printf("keyboard=[%s]\n", buff );
		}

		if( FD_ISSET(fd, &readfds) )
		{
			ret = read(fd, buff, sizeof buff );
			buff[ret-1] = 0;
			printf("myfifo  =[%s]\n", buff );
		}
	}
	return 0;
}















