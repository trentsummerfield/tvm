public class Printer {
	public void printMsg() {
		print("Hello from printer\n");
	}

	public static native void print(String s);
}
