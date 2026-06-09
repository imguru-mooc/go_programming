#include <stdio.h>
#include <fcntl.h>
#include <errno.h>
#include <string.h>

#if 1

int main()
{
	int fd;
	fd = open("hello.txt", O_RDONLY);

    if( fd < 0 )
    {
       printf("strerror=%s\n", strerror(errno) );
       return -1;
    }
}
#endif


#if 0
int divide(int a, int b, int *remainder) {
    *remainder = a % b;
    return a / b;
}

int main()
{
    int remainder;
    int result;
    result = divide( 10, 3, &remainder);
    printf("result=%d, remainder=%d\n", result, remainder);
}
#endif

#if 0
// C
int add(int a, int b) {
    return a + b;
}

int main()
{
	printf("%d\n" , add(1,2));
	return 0;
}
#endif
