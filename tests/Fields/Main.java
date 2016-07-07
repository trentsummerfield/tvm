public class Main {
	public static void main(String[] args) {
		Child object = new Child();
		println(object.fieldInChild);
		println(object.fieldInParent);
		println(object.fieldOverriddenInChild);
		println(object.arrayField);
	}

	public static void println(Object s) {
		if (s == null) {
			print("<null>\n");
		} else if (s instanceof String) {
			print(s + "\n");
		} else if (s instanceof Object[]) {
			print("Object array\n");
		}
	}

	public static native void print(String s);
}
