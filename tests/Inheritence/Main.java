public class Main {
    public static void main(String[] args) {
        Cat cat = new Cat();
        Dog dog = new Dog();
        Dog angryDog = new BarkingDog();

        talkTo(cat);
        talkTo(dog);
        talkTo(angryDog);
    }

    public static void talkTo(Animal a) {
        print("Hello " + a.name() + "\n");
        print(a.talk() + "\n");
    }

    public static native void print(String s);
}
