import * as EV from '../evalvite';
import * as S5Conf from '../conf';
import * as S5Render from '../render';
import rowColBase, { flatRowColInputs } from '../render/rowcolbase';
import base from './base';

let { IdCounter } = S5Conf.Vars;
const { Debug } = S5Conf.Vars;

export default class row extends rowColBase {
  constructor(debugName?: string) {
    super(debugName || `[row ${IdCounter}]`);
    if (!debugName) {
      IdCounter += 1;
    }

    // replace the width constraint on this interactor
    this.Coords.AttrW.SetInputsAndFunc(
      [this.parts, this._margin],
      (f: flatRowColInputs[], margin: number): number => {
        // special case
        if (f.length === 0) {
          return 10;
        }
        // normal case
        let width = 0; // for the same as the element of the height below, get H
        let maxX = 0; // corresponds to max X of the models
        f.forEach((m: flatRowColInputs) => {
          if (m.x > maxX) {
            maxX = m.x;
            width = m.w;
          }
        });
        return maxX + width + margin;
      }
    );
    // replace the height constraint on this interactor
    this.Coords.AttrH.SetInputsAndFunc(
      [this.parts, this._margin],
      (f: flatRowColInputs[], margin: number): number => {
        // special case
        if (f.length === 0) {
          return 20;
        }
        // normal case
        let maxH = 0; // largest height
        f.forEach((m: flatRowColInputs) => {
          if (m.h > maxH) {
            maxH = m.h;
          }
        });
        return maxH + 2 * margin;
      }
    );
  }

  RenderSelf(ctx: S5Render.Context): void {
    if (!this._transparentBackground) {
      ctx.FillStyle = S5Render.DefaultTheme.Background;
      ctx.FillRect(0, 0, this.W, this.H);
    }
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types,@typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars,class-methods-use-this
  UpdateValue(a: EV.Attribute<any>, v?: any, old?: any) {
    if (Debug) {
      S5Conf.Log(`S5DEBUG: updated value on row: ${a.DebugName()} to ${v}`);
    }
  }

  addLayoutConstraints(di: base): void {
    // all the Y values are === margin
    di.Coords.AttrY.SetInputsAndFunc(
      [this._margin],
      (margin: number): number => {
        return margin;
      }
    );

    // x value changes if it is the first element or not
    if (this.Children.length === 0) {
      di.Coords.AttrX.SetInputsAndFunc(
        [this._margin],
        (margin: number): number => {
          return margin;
        }
      );
    } else {
      const prev = this.Children[this.Children.length - 1];
      // xxx fixme(iansmith) cast
      const pi = prev as base;
      di.Coords.AttrX.SetInputsAndFunc(
        [pi.Coords.AttrX, pi.Coords.AttrW, this._margin],
        (prevX: number, prevW: number, margin: number): number => {
          return prevX + prevW + margin;
        }
      );
    }
  }

  // only need to remove the ones on di, because the array parts is
  // handled by superclass
  removeLayoutConstraints(di: base): void {
    di.Coords.AttrX.DropAttributeConnection(this._margin);
    di.Coords.AttrY.DropAttributeConnection(this._margin);

    const i = base.ChildIndex(di); // could be "interactor"
    if (i > 0) {
      // fixme(iansmith)
      const prev = this.Children[i - 1] as base;
      di.Coords.AttrX.DropAttributeConnection(
        prev.Coords.AttrX,
        prev.Coords.AttrW
      );
    }
  }
}
