package java.lang;

public class StringBuilder {
    private char[] data;
    private int size;

    public StringBuilder() {
        data = new char[8];
        size = 0;
    }

    public StringBuilder append(String s) {
        int l = s.value.length;
        while (size + l > data.length) {
            char[] expandedData = new char[data.length * 2];
            for (int i = 0; i < size; i++) {
                expandedData[i] = data[i];
            }
            data = expandedData;
        }
        for (int i = 0; i < l; i++) {
            data[i+size] = s.value[i];
        }
        size += l;
        return this;
    }

    public String toString() {
        return new String(data, size);
    }
}
