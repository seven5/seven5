import * as EV from '../evalvite';
import * as S5Conf from '../conf';
import rowColBase, {
  flatRowColInputs,
  rowColInputs,
} from '../render/rowcolbase';
import base from './base';
import { Bind } from '../evalvite';

let { IdCounter } = S5Conf.Vars;
const { Debug } = S5Conf.Vars;

export default class column extends rowColBase {
  constructor(debugName?: string) {
    super(debugName || `[column ${IdCounter}]`);
    if (!debugName) {
      IdCounter += 1;
    }

    // replace the height constraint on this interactor
    this.Coords.AttrH.SetInputsAndFunc(
      [this.parts, this._margin],
      (f: flatRowColInputs[], margin: number): number => {
        // special case
        if (f.length === 0) {
          return 20;
        }
        // normal case
        let height = 0; // for the same as the element of the height below, get H
        let maxY = 0; // corresponds to max Y of the models
        f.forEach((m: flatRowColInputs) => {
          if (m.y > maxY) {
            maxY = m.y;
            height = m.h;
          }
        });
        return maxY + height + margin;
      }
    );
    // replace the width constraint on this interactor
    this.Coords.AttrW.SetInputsAndFunc(
      [this.parts, this._margin],
      (f: flatRowColInputs[], margin: number): number => {
        // special case
        if (f.length === 0) {
          return 10;
        }
        // normal case
        let maxW = 0; // largest width
        f.forEach((m: flatRowColInputs) => {
          if (m.w > maxW) {
            maxW = m.w;
          }
        });
        return maxW + 2 * margin;
      }
    );
    Bind<number>(this.Coords.AttrH, this);
    Bind<number>(this.Coords.AttrW, this);
    Bind<rowColInputs[]>(this.parts, this);
  }

  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types,@typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars,class-methods-use-this
  UpdateValue(a: EV.Attribute<any>, v?: any, old?: any) {
    if (Debug) {
      S5Conf.Log(`S5DEBUG: updated value on column: ${a.DebugName()} to ${v}`);
    }
  }

  addLayoutConstraints(di: base): void {
    if (this.Children.length === 0) {
      // bind the location to the margin
      di.Coords.AttrX.SetInputsAndFunc(
        [this._margin],
        (n: number): number => n
      );
      di.Coords.AttrY.SetInputsAndFunc(
        [this._margin],
        (n: number): number => n
      );
    } else {
      const prevIndex = this.Children.length - 1;
      const pi = this.Children[prevIndex] as base;
      // we need to now insure the object *in* the column has
      // the right constraint on it's X and Y... we add two
      // new constraints
      di.Coords.AttrX.SetInputsAndFunc([pi.Coords.AttrX], (x) => x);

      // Y position is the sum of prev element's Y, the prev
      // elements height, and the margin
      di.Coords.AttrY.SetInputsAndFunc(
        [pi.Coords.AttrY, pi.Coords.AttrH, this._margin],
        (yPrev: number, hPrev: number, m: number): number => {
          return yPrev + hPrev + m;
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
      const prev = this.Children[i - 1] as base;
      di.Coords.AttrY.DropAttributeConnection(
        prev.Coords.AttrY,
        prev.Coords.AttrH
      );
    }
  }
}
