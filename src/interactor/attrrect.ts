// needs to precede interactor
import * as EV from '../evalvite';
import * as S5Render from '../render';
import * as S5Err from '../error';

export default class attrRect {
  AttrX: EV.Attribute<number>;

  AttrY: EV.Attribute<number>;

  AttrW: EV.Attribute<number>;

  AttrH: EV.Attribute<number>;

  constructor(interactor?: EV.Updateable<number>) {
    this.AttrW = EV.Simple<number>(40, 'attrW');
    this.AttrH = EV.Simple<number>(20, 'attrH');
    this.AttrX = EV.Simple<number>(10, 'attrX');
    this.AttrY = EV.Simple<number>(30, 'attrY');
    if (interactor) {
      this.AttrX.AddUpdateable(interactor);
      this.AttrY.AddUpdateable(interactor);
      this.AttrW.AddUpdateable(interactor);
      this.AttrH.AddUpdateable(interactor);
    }
  }

  nameToTarget(n: S5Render.CoordName): EV.Attribute<number> {
    let target: EV.Attribute<number> | null;
    switch (n) {
      case S5Render.CoordName.X:
        target = this.AttrX;
        break;
      case S5Render.CoordName.Y:
        target = this.AttrY;
        break;
      case S5Render.CoordName.W:
        target = this.AttrW;
        break;
      case S5Render.CoordName.H:
        target = this.AttrH;
        break;
      default:
        // this really shouldn't happen, it means we didn't cover all
        // the cases in the switch, which is an enum so...
        throw S5Err.NewError(S5Err.Messages.NotImplemented, `coord {$n}`);
    }
    if (target === null) {
      // this really shouldn't happen, the compiler cant decide that
      // the switch above always assigns to target
      throw S5Err.NewError(S5Err.Messages.NotImplemented, `coord {$n}`);
    }
    return target;
  }

  Get(n: S5Render.CoordName): number {
    const target = this.nameToTarget(n);
    if (target === null) {
      throw S5Err.NewError(S5Err.Messages.NullAttr, `bad dimension ${n}`);
    }
    return target.Get();
  }

  Set(n: S5Render.CoordName, v: number): void {
    const target = this.nameToTarget(n);
    if (target !== null) {
      if (target.AllowsSet()) {
        target.Set(v);
      } else {
        throw S5Err.NewError(S5Err.Messages.Constrained, `${n}`);
      }
    } else {
      throw S5Err.NewError(S5Err.Messages.NullAttr, `bad dimension ${n}`);
    }
  }

  rect(): S5Render.Rect {
    return new S5Render.Rect(
      this.Get(S5Render.CoordName.X),
      this.Get(S5Render.CoordName.Y),
      this.Get(S5Render.CoordName.W),
      this.Get(S5Render.CoordName.H)
    );
  }
}
