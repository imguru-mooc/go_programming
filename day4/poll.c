#include <unistd.h>
#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/select.h>
#include <poll.h>

int main()
{
	int ret;
	int fd;
	char buff[1024];

	struct pollfd fds[2] = {0,};

	fd = open("myfifo", O_RDWR);  // mkfifo myfifo

	while(1)
	{
		fds[0].fd = 0; 
		fds[0].events = POLLIN;

		fds[1].fd = fd; 
		fds[1].events = POLLIN;

		poll(fds, 2, 1000);

		if( fds[0].revents & POLLIN )
		{
			ret = read(0, buff, sizeof buff );
			buff[ret-1] = 0;
			printf("keyboard=[%s]\n", buff );
		}

		if( fds[1].revents & POLLIN )
		{
			ret = read(fd, buff, sizeof buff );
			buff[ret-1] = 0;
			printf("myfifo  =[%s]\n", buff );
		}
	}
	return 0;
}















