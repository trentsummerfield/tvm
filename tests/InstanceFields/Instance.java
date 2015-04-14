public class Instance {
    private int x;

    public void printField() {
	printInt(x);
    }

    public void setField(int y) {
	x = y;
    }

    public static native void printInt(int i);
}
