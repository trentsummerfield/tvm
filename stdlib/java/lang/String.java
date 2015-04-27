package java.lang;

public class String {
    public char[] data;

    public String(char[] characters, int length) {
        data = new char[length];
        for (int i = 0; i < length; i++) {
            data[i] = characters[i];
        }
    }
}
