public class IntegerArithmetic {
	private static int FIFTY_THOUSAND = 50_000;

	public static void main(String[] args) {
		printInt(add(5, 15));
		printInt(minus(30, 2));
		printInt(minus(55, 60));
		printInt(mult(2, FIFTY_THOUSAND));
		printInt(div(20, 10));
	}

	public static int add(int x, int y) {
		return x + y;
	}

	public static int minus(int x, int y) {
		return x - y;
	}

	public static int mult(int x, int y) {
		return x * y;
	}

	public static int div(int x, int y) {
		return x / y;
	}

	public static native void printInt(int x);
}
