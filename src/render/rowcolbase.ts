// sort of a "model" of each of the members of the column, each instance
// of this model represents the contribution of a member to the H of the column
// (h and y) and the W of the column
import { Attribute } from '../evalvite';
import * as EV from '../evalvite';
import base from '../interactor/base';
import interactor from '../interactor/interactor';
import * as S5Err from '../error';
import arrayAttribute from '../evalvite/arrayattr';
import * as S5Render from './index';
import attrPrivateImpl from '../evalvite/attrprivate';

export interface rowColInputs {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
  h: Attribute<number>;
  y: Attribute<number>;
  w: Attribute<number>;
  x: Attribute<number>;
}

// the model for computing the correct size(s) of a row or column
export type flatRowColInputs = {
  h: number;
  y: number;
  w: number;
  x: number;
};

export default abstract class rowColBase extends base {
  _margin: EV.Attribute<number> = EV.Simple<number>(2, 'col margin');

  // should this return the attribute or the value of it?
  get Margin(): number {
    return this._margin.Get();
  }

  set Margin(m: number) {
    this._margin.Set(m);
    this.MarkDirty();
  }

  // this is the crucial array of the models that represent the height/pos
  // of each member of the column
  parts: EV.Attribute<rowColInputs[]> = EV.ArrayAttr<rowColInputs>(
    'parts_of_row_or_column'
  );

  Render(ctx: S5Render.Context): void {
    (this.Coords.AttrW as attrPrivateImpl<number>).markDirty();
    (this.Coords.AttrH as attrPrivateImpl<number>).markDirty();
    super.Render(ctx);
  }

  RenderSelf(ctx: S5Render.Context): void {
    if (!this._transparentBackground) {
      ctx.FillStyle = S5Render.DefaultTheme.Background;
      ctx.FillRect(0, 0, this.W, this.H);
    }
  }

  // eslint-disable-next-line class-methods-use-this,@typescript-eslint/no-unused-vars
  AddChildAt(i: interactor, x: number, y: number): void {
    throw S5Err.NewError(
      S5Err.Messages.Constrained,
      'position of child in row or column cannot be set'
    );
  }

  AddChild(i: interactor): void {
    // ugh, had to do a cast to get access to the attributes fixme(iansmith)
    const di = i as base;
    const model: rowColInputs = {
      h: di.Coords.AttrH,
      y: di.Coords.AttrY,
      w: di.Coords.AttrW,
      x: di.Coords.AttrX,
    };
    this.addLayoutConstraints(di);
    super.AddChild(di);

    // xxx fixme(iansmith) cast to the array type
    // array type does the correct marking when you do push
    (this.parts as arrayAttribute<rowColInputs>).push(model);
  }

  RemoveChild(i: interactor): boolean {
    const di = i as base;
    const model: rowColInputs = {
      h: di.Coords.AttrH,
      y: di.Coords.AttrY,
      w: di.Coords.AttrW,
      x: di.Coords.AttrX,
    };

    // first, figure out which child is the target or error if not found
    const targetIndex = base.ChildIndex(i, true);
    if (targetIndex < 0) {
      return false;
    }

    // second, confirm that the same index of the MODELS is the one for this interactor
    // fixme(iansmith) this cast happens a lot with array attributes
    const a = this.parts as arrayAttribute<rowColInputs>;
    const candModel = a.Raw()[targetIndex];
    if (
      !(
        candModel.x === model.x &&
        candModel.y === model.y &&
        candModel.w === model.w &&
        candModel.h === model.h
      )
    ) {
      throw S5Err.NewError(
        S5Err.Messages.BadState,
        'RemoveChild(): found child, but not its corresponding model'
      );
    }
    // now, with everything confirmed, disconnect constraints
    this.removeLayoutConstraints(di);
    // xxx fixme(iansmith) cast to the array type
    // array type does the correct unmarking when you do splice
    (this.parts as arrayAttribute<rowColInputs>).splice(targetIndex, 1);
    return super.RemoveChild(i);
  }

  // xxx fixme(iansmith) really should be passing interactor, not defaultinteractor here
  abstract removeLayoutConstraints(di: base): void;

  abstract addLayoutConstraints(di: base): void;
}
