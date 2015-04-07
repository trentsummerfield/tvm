public class IntegerArithmetic {
	public static void main(String[] args) {
		printInt(add(5, 15));
		printInt(minus(30, 2));
		printInt(minus(55, 60));
	}

	public static int add(int x, int y) {
		return x + y;
	}

	public static int minus(int x, int y) {
		return x - y;
	}

	public static native void printInt(int x);
}
