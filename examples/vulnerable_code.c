// Example vulnerable C code - Buffer Overflow
// CWE-120: Buffer Copy without Checking Size of Input ('Classic Buffer Overflow')

#include <string.h>
#include <stdio.h>

void copy_user_input(char *user_input)
{
    char buffer[64];

    // VULNERABLE: No bounds checking on strcpy
    // This can cause buffer overflow if user_input > 64 bytes
    strcpy(buffer, user_input);

    printf("Copied: %s\n", buffer);
}

int main()
{
    char input[256];
    printf("Enter data: ");
    fgets(input, sizeof(input), stdin);

    copy_user_input(input);

    return 0;
}
