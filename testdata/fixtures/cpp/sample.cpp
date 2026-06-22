int free_function(int x, int y) {
    return x + y;
}

class Calculator {
public:
    int add(int a, int b);
    int divide(int a, int b, int default_val = 0);
};

int Calculator::add(int a, int b) {
    return a + b;
}

int Calculator::divide(int a, int b, int default_val) {
    if (b == 0) return default_val;
    return a / b;
}

struct Point {
    int x;
    int y;
};

inline int max(int a, int b) {
    return (a > b) ? a : b;
}

int process_data(const int &value) {
    return value * 2;
}
