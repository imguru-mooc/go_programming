#if 1
#include <stdio.h>
class CAR
{
	public:
		virtual void run()=0;
};

class MATIZ : public CAR
{
	public:
		virtual void run() { printf("마티즈가 달린다\n");}
};

int main()
{
	CAR *p;
	p = new(MATIZ);
	p->run();
}
#endif
#if 0
#include <stdio.h>
class CAR
{
	public:
		virtual void run() { printf("자동차가 달린다\n");}
};

class MATIZ : public CAR
{
	public:
		virtual void run() { printf("마티즈가 달린다\n");}
};

int main()
{
	CAR *p;
	p = new(MATIZ);
	p->run();
}
#endif

#if 0
#include <stdio.h>
class CAR
{
	public:
		void run() { printf("자동차가 달린다\n");}
};

class MATIZ : public CAR
{
	public:
		void run() { printf("마티즈가 달린다\n");}
};

int main()
{
	MATIZ m;
	m.run();
}
#endif

#if 0
#include <stdio.h>
class CAR
{
	public:
		void run() { printf("자동차가 달린다\n");}
};

int main()
{
	CAR car;
	car.run();
}
#endif
