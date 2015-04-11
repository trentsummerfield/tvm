public class StaticField {
    public static int field = 101;

    public static void main(String[] args) {
		printInt(field);
    }

    public static native void printInt(int i);
}
