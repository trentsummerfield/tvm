public class MethodCalls {
    public static void main(String[] args) {
	sayHello();
	sayWorld();
    }

    public static void sayHello() {
	print("Hello\n");
    }

    public static void sayWorld() {
	print("World\n");
    }

    public static native void print(String s);
}
