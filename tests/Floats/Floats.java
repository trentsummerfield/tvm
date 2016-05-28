public class Floats {
	public static void main(String[] args) {
		printFloat(add(5.1f, 15.4f));
		printFloat(minus(30.2f, 2.f));
		printFloat(minus(55.1f, 60.1f));
		printFloat(mult(2.1f, 50.f));
		printFloat(div(5.0f, 2.0f));
	}

	public static float add(float x, float y) {
		return x + y;
	}

	public static float minus(float x, float y) {
		return x - y;
	}

	public static float mult(float x, float y) {
		return x * y;
	}

	public static float div(float x, float y) {
		return x / y;
	}

	public static native void printFloat(float x);
}
