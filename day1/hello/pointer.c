#include <stdio.h>

#if 1
int main()
{
	int a[2][2] = {{1, 2}, {3, 4}};
	int (*p)[2]  = a;

	p[1][1] = 10;
}
#endif
#if 0
int main()
{
	int a[2][2] = {{1, 2}, {3, 4}};
	int *p = a;

	p[1][1] = 10;
}
#endif
#if 0
int main()
{
	char a = 100;
	int  *i = &a;
	printf("i = %d\n", *i); 
}
#endif

#if 0
int main()
{
	int a[4] = {1, 2, 3, 4};
	int *p = a;
	printf("a = %lu\n", sizeof(a)); 
	printf("a = %lu\n", sizeof(int [4])); 
	printf("p = %lu\n", sizeof(p)); 
	printf("p = %lu\n", sizeof(int *)); 
}
#endif

#if 0
int main()
{
	// C — 가능 (위험하지만)
	int arr[5] = {1, 2, 3, 4, 5};
	int *p = arr;
	p++;         // OK — 다음 원소로 이동
	printf("%d\n", *p);  // 2
}
#endif

#if 0
int main()
{
    int x = 10;
    int *p = &x;
    *p = 20;

    printf("x=%d\n", x);
}
#endif
