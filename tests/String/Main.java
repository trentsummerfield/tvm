public class Main {
    public static String hello = "Hello";
    public static String world = "World";
    public static String space = " ";

    public static void main(String[] args) {
        print(hello + space + world);
    }

    public static native void print(String s);
}
