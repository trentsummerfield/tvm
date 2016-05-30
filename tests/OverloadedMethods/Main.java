public class Main {
	private static String str = "A string";
	private static Object o = new Object();

	public static void main(String[] args) {
		type(str);
		type(o);
	}

	public static void type(Object o) {
		print("Called with an Object\n");
	}

	public static void type(String s) {
		print("Called with a String\n");
	}

	public static native void print(String s);
}
