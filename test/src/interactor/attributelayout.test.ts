import {expect} from 'chai';
import * as S5Interactor from "../../../src/interactor";

describe("attribute based layout", ()=> {
  const toTheRight = 50;
  let r: S5Interactor.Root;
  let i1: S5Interactor.Base;
  let i2: S5Interactor.Base;
  const column=new S5Interactor.Column('test_col');
  const row=new S5Interactor.Row('test_row');
  const newWidth = 92;

  before(() => {
    r = new S5Interactor.Root();
  });

  beforeEach(() => {
    r.ClearChildren();
    column.ClearChildren();
    row.ClearChildren();
    i1 = new S5Interactor.Base("i1");
    i2 = new S5Interactor.Base("i2");

  });

  it('should put i2 to the right of i1', () => {
    i2.Coords.AttrY.SetInputsAndFunc([i1.Coords.AttrY],(y)=>y);
    i2.Coords.AttrX.SetInputsAndFunc(
      [i1.Coords.AttrX, i1.Coords.AttrW],
      (x: number, w: number): number => x + w + toTheRight,
    );

    r.AddChildAt(i1, 100, 10);
    r.AddChild(i2);

    r.Mainloop();

    expect(i2.X - toTheRight).to.be.equal(i1.X + i1.W);
    expect(i1.Y).to.be.equal(i2.Y);
  });

  it('should place i2 below i1 after swap of layout constraint', () => {
    i2.Coords.AttrY.SetInputsAndFunc([i1.Coords.AttrY],
    (p: number): number => {
      return p + 10;// equal by another means
    });

    r.Mainloop();

    expect(i1.Y + 10).to.be.equal(i2.Y);

  });
  it('should adapt column to changing size of inputs', () => {
    const checkI2YPosAndColHeight = (pre:string) => {
      expect(i2.Y).to.be.equal(i1.Y+i1.H+column.Margin,`${pre} i2.Y`);
      expect(column.H).to.be.equal(i2.Y+i2.H+column.Margin, `${pre} column.H`);
    }

    // clear out the root
    r.RemoveChild(i1);
    r.RemoveChild(i2);

    column.AddChild(i1);
    column.AddChild(i2);

    //put column in root
    r.AddChildAt(column, 40,40);

    // test with default margin
    r.Mainloop();
    checkI2YPosAndColHeight('normal margin');

    //change margin
    column.Margin=0;
    r.Mainloop();
    checkI2YPosAndColHeight('adjusted margin');

    // put it back to 2
    column.Margin=2;

    // change height of early object in col
    const sizeIncrement = 20;
    const prevI1H=i1.H;
    const prevI2Y=i2.Y;
    i1.H = prevI1H+sizeIncrement;
    r.Mainloop();

    // this check prevents a possible bug where the column just ignores the
    // size change on i1... that would pass checkI2YPosAndColHeight
    expect(i2.Y).to.be.equal(prevI2Y+sizeIncrement,'Y2 should have moved');

    checkI2YPosAndColHeight('adjusted height of i1');

    i2.W = newWidth;
    r.Mainloop();
    expect(column.W).to.be.equal(2*column.Margin + newWidth);
  });
  it('should adapt row to changing size of inputs', () => {

    //sanity check the cleanup code in beforeEach
    expect(r.Children.length).to.be.equal(0);
    expect(row.Children.length).to.be.equal(0);
    expect(column.Children.length).to.be.equal(0);

    row.AddChild(i1);
    row.AddChild(i2);

    r.AddChildAt(row,40,40);

    r.Mainloop();
    // 3 because left margin, right margin, and gap between two items
    expect(row.W).to.be.equal(3*row.Margin + i2.W + i1.W);

    row.RemoveChild(i2);
    // 2 because left margin, right margin
    expect(row.W).to.be.equal(2*row.Margin + i1.W);

  });
});