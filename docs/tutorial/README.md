# Tutorial 🤓

## Install the tools

If you have a linux machine you can run the following command: 

```bash
curl https://raw.githubusercontent.com/0xleft/gal/main/install.sh | sudo bash
```

If you are developing on windows you must download the installer from Download installer from [here](https://github.com/0xleft/gal/releases/latest/download/gal_installer.exe) and run it.

## Hello world! 

After you have installed the tooling you can test them by running `gal` or `gal.exe` on windows.

Now that we are ready to do some coding lets start of with a clasic "Hello World" script.

1. Create a file called hello_world.gal using your favorite editor
2. Create a function "main" as this is where our program will start executing from

```gal
lowkey main{}
    ` our function content will go here
end
```

3. The standart library includes a function called `std.print` we can use it to print a string like the following program.

To call the function we use the `fire` keyword.

```gal
lowkey main{}
    fire std.print("Hello World!")
end
```

4. Execute this script in your terminal using `gal run hello_world.gal`

## Variables 🤑

Variables are defined using the `fax keyword`.

```gal
lowkey main{}
    fax a = 2 ` here we define variable a with value of 2
    a = 3 ` after variable a has been declared we can assign values without using the fax keyword
end
```

## Comments 😤

Anything after ` symbol will be a comment and ignored in the execution of the script.

## Loops and if statements

Loops are called `durin` like "during" you get it 😉, and if statements are called `foreal` like `for real?`. I must stop with the puns.

```gal
lowkey main{}
    fax a = 3
    foreal a > 2
        std.print("A is more than 2")
    end
end
```

```gal
lowkey main{}
    fax a = 3 ` define the variable a
    durin a < 100 ` while a is less than 100
        std.println(a) ` we are using std.println as it ads an extra line after each print.
        a = a + 1 ` add one to the variable a
    end
end
```

## Lists and dictionaries

In gal (genalphalang) any variable can be a dictionary or a list for example:

```gal
lowkey main{}
    fax long_string = "h,e,l,l,o, ,w,o,r,l,d"
    ` we can use a std.split to split it up
    fax separated = fire std.split(long_string, ",")
    fax i = 0
    ` print the "hello world"
    durin [separated i] != nuthin
        fire std.print([separated i])
        i = i + 1
    end
end
```

We can use strings for indecies too

```gal
lowkey main{}
    fax number = 1
    [number "test"] = 1
    fire std.println([number "test"])
    ` this script will print "1"
end
```

## The end

This is the end of this short tutorial if you would like more info make sure to ask in issues or discord