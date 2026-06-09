#if 1
#include <stdio.h>

struct ST
{
	int kor, eng, math;
	int total;
	double aver;

	void input();
	void calc();
	void output();
};


int main()
{
	ST s;
	s.input(); 
	s.calc();
	s.output();

	return 0;
}

void ST::input()
{
	kor=89;
	eng=90;
	math=99;
}

void ST::calc()
{
	total = kor + eng + math;
	aver = total / 3.;
}

void ST::output()
{
	printf("kor=%d, eng=%d, math=%d, total=%d, aver=%5.2lf\n", 
			kor, eng, math, total, aver);
}

#endif

#if 0
#include <stdio.h>

typedef struct
{
	int kor, eng, math;
	int total;
	double aver;
} ST;

void input( ST *s );
void calc( ST *s );
void output( ST *s );

int main()
{
	ST s;
	input( &s ); 
	calc( &s );
	output( &s );

	return 0;
}

void input( ST *s )
{
	s->kor=89;
	s->eng=90;
	s->math=99;
}

void calc(ST *s)
{
	s->total = s->kor + s->eng + s->math;
	s->aver = s->total / 3.;
}

void output(ST *s)
{
	printf("kor=%d, eng=%d, math=%d, total=%d, aver=%5.2lf\n", 
			s->kor, s->eng, s->math, s->total, s->aver);
}

#endif
#if 0
#include <stdio.h>

void input(int *kor, int *eng, int *math);
void calc(int kor, int eng, int math, int *total, double *aver);
void output(int kor, int eng, int math, int total, double aver);

int main()
{
	int kor, eng, math;
	int total;
	double aver;

	input( &kor, &eng, &math ); 
	calc( kor, eng, math, &total, &aver );
	output(kor, eng, math, total, aver );

	return 0;
}

void input(int *kor, int *eng, int *math)
{
	*kor=89;
	*eng=90;
	*math=99;
}

void calc(int kor, int eng, int math, int *total, double *aver)
{
	*total = kor + eng + math;
	*aver = *total / 3.;
}

void output(int kor, int eng, int math, int total, double aver)
{
	printf("kor=%d, eng=%d, math=%d, total=%d, aver=%5.2lf\n", 
			kor, eng, math, total, aver);
}

#endif
#if 0
#include <stdio.h>

void input(int *kor, int *eng);
void calc(int kor, int eng, int *total, double *aver);
void output(int kor, int eng, int total, double aver);

int main()
{
	int kor, eng;
	int total;
	double aver;

	input( &kor, &eng ); 
	calc( kor, eng, &total, &aver );
	output(kor, eng, total, aver );

	return 0;
}

void input(int *kor, int *eng)
{
	*kor=89;
	*eng=90;
}

void calc(int kor, int eng, int *total, double *aver)
{
	*total = kor + eng;
	*aver = *total / 2.;
}

void output(int kor, int eng, int total, double aver)
{
	printf("kor=%d, eng=%d, total=%d, aver=%5.2lf\n", 
			kor, eng, total, aver);
}

#endif
#if 0
#include <stdio.h>
int main()
{
	int kor=23, eng=34;
	int total;
	double aver;

	total = kor + eng;
	aver = total / 2.;

	printf("kor=%d, eng=%d, total=%d, aver=%5.2lf\n", 
			kor, eng, total, aver);

	return 0;
}
#endif
