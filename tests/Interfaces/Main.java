public class Main {
	public static void main(String[] args) {
		Cat cat = new Cat();
		Dog dog = new Dog();
		sayHello(cat);
		sayHello(dog);
	}

	public static void sayHello(Nameable thing) {
		print("Hello " + thing.name() + "\n");
	}

	private static native void print(String s);
}
