package java.lang;

public class String {
    public char[] value;

    public String(char[] characters, int length) {
        value = new char[length];
        for (int i = 0; i < length; i++) {
            value[i] = characters[i];
        }
    }
}
