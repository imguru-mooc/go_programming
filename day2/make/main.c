#include <stdio.h>
#include <stdlib.h>
#include "main.h"

int main()
{
	char *s = "1234";
	int n;


	n = my_atoi(s);
	n++;
	printf("%d\n", n );
	return 0;
}


