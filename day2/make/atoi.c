#include "main.h"

int my_atoi( char *s )
{
   int n=0;
   int i;


   for(i=0; s[i] ; i++)
   {
       n = n*10 + s[i] - '0';
   }
   return n;
}
