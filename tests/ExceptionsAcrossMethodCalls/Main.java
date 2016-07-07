public class Main {
	public static void main(String[] args) throws Throwable {
		try {
			thrower(new IllegalStateException());
		} catch (Exception e) {
			print("Shouldn't get here\n");
		} finally {
			print("Finally in main\n");
		}

		try {
			thrower(new java.io.IOException());
			print("Shouldn't get here\n");
		} catch (Exception e) {
			print("Caught exception in main\n");
		} finally {
			print("Finally in main\n");
		}
	}

	private static void thrower(Throwable t) throws Throwable {
		try {
			throw t;
		} catch (RuntimeException e) {
			print("Caught exception in thrower\n");
		} finally {
			print("Finally in thrower\n");
		}
	}

	public static native void print(String s);
}
