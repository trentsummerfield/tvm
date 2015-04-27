package java.lang;

public class StringBuilder {
    private char[] data;
    private int size;

    public StringBuilder() {
        data = new char[100];
        size = 0;
    }

    public StringBuilder append(String s) {
        int l = s.data.length;
        for (int i = 0; i < l; i++) {
            data[i+size] = s.data[i];
        }
        size += l;
        return this;
    }

    public String toString() {
        return new String(data, size);
    }
}
