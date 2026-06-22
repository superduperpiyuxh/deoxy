int add(int a, int b) {
    return a + b;
}

void process_buffer(char *buffer, int size) {
    if (buffer != NULL) {
        buffer[0] = 'A';
    }
}

void log_message(const char *message);

static int internal_helper(int value) {
    return value * 2;
}

int sum_array(int arr[], int count) {
    int total = 0;
    for (int i = 0; i < count; i++) {
        total += arr[i];
    }
    return total;
}
