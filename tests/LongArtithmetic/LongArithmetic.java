public class LongArithmetic {
	private static long FIVE = 5;

	public static void main(String[] args) {
		printLong(add(FIVE, 15));
		printLong(minus(30, 2));
		printLong(minus(55, 60));
		printLong(mult(2, 50));
		printLong(div(20, 10));
	}

	public static long add(long x, long y) {
		return x + y;
	}

	public static long minus(long x, long y) {
		return x - y;
	}

	public static long mult(long x, long y) {
		return x * y;
	}

	public static long div(long x, long y) {
		return x / y;
	}

	public static native void printLong(long x);
}
