public class Main {
    public static void main(String[] args) {
        char[] chars = new char[11];
        chars[0] = 'h';
        chars[1] = 'e';
        chars[2] = 'l';
        chars[3] = 'l';
        chars[4] = 'o';
        chars[5] = ' ';
        chars[6] = 'w';
        chars[7] = 'o';
        chars[8] = 'r';
        chars[9] = 'l';
        chars[10] = 'd';
        printChar(chars[0]);
        printChar(chars[1]);
        printChar(chars[2]);
        printChar(chars[3]);
        printChar(chars[4]);
        printChar(chars[5]);
        printChar(chars[6]);
        printChar(chars[7]);
        printChar(chars[8]);
        printChar(chars[9]);
        printChar(chars[10]);
    }

    public static native void printChar(char c);
}
