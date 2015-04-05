public class IntegerArithmetic {
	public static void main(String[] args) {
		int z = add(5, 15);
		printInt(z);
	}

	public static int add(int x, int y) {
		return x + y;
	}

	public static native void printInt(int x);
}
