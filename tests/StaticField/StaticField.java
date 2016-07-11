public class StaticField {
    public static int field = 101;

    public static void main(String[] args) {
		printInt(field);
		field = 100;
		printInt(field);
    }

    public static native void printInt(int i);
}
