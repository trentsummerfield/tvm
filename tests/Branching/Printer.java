public class Printer {
	private boolean finished;

	public void printMsg() {
		if (!finished) {
			print("Hello\n");
		} else {
			print("Goodbye\n");
		}
	}

	public void finished() {
		finished = true;
	}

	public static native void print(String s);
}
