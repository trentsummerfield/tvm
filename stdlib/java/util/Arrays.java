package java.util;

public class Arrays {
	public static char[] copyOfRange(char[] original, int from, int to) {
		int newLength = to - from;
		char[] copy = new char[newLength];
		int length = original.length;
		if (newLength < length) {
			length = newLength;
		}
		System.arraycopy(original, from, copy, 0, length);
		return copy;
	}
}
