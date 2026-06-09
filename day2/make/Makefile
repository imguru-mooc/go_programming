OBJS = main.o atoi.o
TARGET = myapp
CC = gcc
CFLAGS = -c

$(TARGET) : $(OBJS) 
	$(CC) $(OBJS) -o $(TARGET)

.c.o : 
	$(CC) $(CFLAGS) $< 

clean : 
	rm -rf $(TARGET) $(OBJS)
