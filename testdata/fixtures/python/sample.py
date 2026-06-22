def greet(name: str, age: int) -> str:
    return f"Hello {name}, age {age}"

def do_nothing():
    pass

def add(a: int, b: int) -> int:
    return a + b

class Calculator:
    def __init__(self, brand: str):
        self.brand = brand

    def add(self, a: int, b: int) -> int:
        return a + b

    @classmethod
    def create_default(cls) -> "Calculator":
        return cls("generic")

    def process_args(self, *args: int, **kwargs: str) -> None:
        pass
