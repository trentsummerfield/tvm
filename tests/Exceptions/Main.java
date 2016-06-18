public class Main {
	private static int n = 100;
	public static void main(String[] args) {
		try {
			if (n == 100) {
				throw new RuntimeException();
			}
			print("Shouldn't get here\n");
		} catch (Exception e) {
			print("Caught exception in main\n");
		}
	}

	public static native void print(String s);
}
