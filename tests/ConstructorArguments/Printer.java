public class Printer {
	private String str;

	public Printer(String msg) {
		str = msg;
	}

	public void printMsg() {
		print(str);
	}

	public static native void print(String s);
}
