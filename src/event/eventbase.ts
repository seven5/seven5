import s5event, { pickList } from './s5event';
import * as S5Err from '../error';

export default class eventBase implements s5event {
  _x: number | undefined = undefined;

  _y: number | undefined = undefined;

  _type: string;

  _pickList: pickList | null;

  _raw: Event | null; // HTML level event, we wrap this

  get Type(): string {
    return this._type;
  }

  get X(): number {
    if (this._x === undefined) {
      throw S5Err.NewError(
        S5Err.Messages.BadState,
        "attempt to use X coordinate of event that doesn't have one"
      );
    }
    return this._x;
  }

  get Y(): number {
    if (this._y === undefined) {
      throw S5Err.NewError(
        S5Err.Messages.BadState,
        "attempt to use Y coordinate of event that doesn't have one"
      );
    }
    return this._y;
  }

  get PickList(): pickList | null {
    return this._pickList;
  }

  set PickList(p: pickList | null) {
    this._pickList = p;
  }

  get Raw(): Event | null {
    return this._raw;
  }

  // this function is largely here to avoid creating setters for
  // X and Y.  With no setter for X and Y some errors are found at
  // compile time.
  Translate(adjustX: number, adjustY: number): void {
    if (this._x === undefined || this._y === undefined) {
      throw S5Err.NewError(
        S5Err.Messages.BadState,
        "attempt to use X,Y coordinates of event that doesn't have them"
      );
    }
    this._x -= adjustX;
    this._y -= adjustY;
  }

  constructor(tName: string, raw: Event | null, x?: number, y?: number) {
    this._type = tName;
    this._pickList = null;
    this._raw = raw;
    if (x && y) {
      this._x = x;
      this._y = y;
    }
  }
}
