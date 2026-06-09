#!/bin/sh
set -x 

gcc -c main.c
gcc -c atoi.c
gcc main.o atoi.o -o main
