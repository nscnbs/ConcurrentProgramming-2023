with Ada.Text_IO;

package Array_Sorter is new Ada.Containers.Array_Sorter(Positive);

procedure Dining_Philosophers is

  type Forks is array(1..5) of Integer;

  protected Dining_Monitor is
    procedure Take_Fork(I: Positive);
    procedure Release_Fork(I: Positive);
  private
    Fork: Forks := (1, 1, 1, 1, 1);
    OK_to_Eat: Vector(1..5) of Array_Sorter.Condition;
  end Dining_Monitor;

  protected body Dining_Monitor is
    procedure Take_Fork(I: Positive) is
    begin
      if Fork(I) /= 2 then
        OK_to_Eat(I).Wait;
      end if;

      Fork((I + 1) mod 5) := Fork((I + 1) mod 5) - 1;
      Fork((I - 1) mod 5) := Fork((I - 1) mod 5) - 1;
    end Take_Fork;

    procedure Release_Fork(I: Positive) is
    begin
      Fork((I + 1) mod 5) := Fork((I + 1) mod 5) + 1;
      Fork((I - 1) mod 5) := Fork((I - 1) mod 5) + 1;

      if Fork((I + 1) mod 5) = 2 then
        OK_to_Eat((I + 1) mod 5).Signal;
      end if;

      if Fork((I - 1) mod 5) = 2 then
        OK_to_Eat((I - 1) mod 5).Signal;
      end if;
    end Release_Fork;
  end Dining_Monitor;

  task type Philosopher(Id: Positive) is
    pragma Priority(Id);
  end Philosopher;

  task body Philosopher is
  begin
    loop
      Think;
      Dining_Monitor.Take_Fork(Id);
      Eat;
      Dining_Monitor.Release_Fork(Id);
    end loop;
  end Philosopher;

  procedure Think is
  begin
    delay Milliseconds(rand(100, 500));
  end Think;

  procedure Eat is
  begin
    Put_Line("Philosopher " & Positive'Image(Id) & " is eating");
    delay Milliseconds(rand(100, 500));
    Put_Line("Philosopher " & Positive'Image(Id) & " finished eating");
  end Eat;

  task type Controller is
  end Controller;

  task body Controller is
    Philosophers: array(1..5) of Philosopher;
  begin
    for I in Philosophers'Range loop
      Philosophers(I) := new Philosopher'(Id => I);
    end loop;

    loop
      delay 1.0;
    end loop;
  end Controller;

begin
  null;
end Dining_Philosophers;
