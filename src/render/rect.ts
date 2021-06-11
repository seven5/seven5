import * as S5Err from '../error';

export enum coordName {
  'X',
  'Y',
  'W',
  'H',
}

export default class rect {
  _x: number;

  _y: number;

  _w: number;

  _h: number;

  constructor(x: number, y: number, w: number, h: number) {
    this._x = x;
    this._y = y;
    this._h = h;
    this._w = w;
  }

  get X(): number {
    return this._x;
  }

  set X(n: number) {
    this._x = n;
  }

  get Y(): number {
    return this._y;
  }

  set Y(n: number) {
    this._y = n;
  }

  get Width(): number {
    return this._w;
  }

  set Width(n: number) {
    this._y = n;
  }

  get Height(): number {
    return this._h;
  }

  set Height(n: number) {
    this._h = n;
  }

  /**
   * X coord of upper left.
   */
  get X1(): number {
    return this._x;
  }

  /**
   * Y coord of upper left.
   */
  get Y1(): number {
    return this._y;
  }

  /**
   * X coord of lower right. Note that a width of zero would not cover any
   * pixels.
   */
  get X2(): number {
    return this._y + this._w;
  }

  /**
   * Y coord of lower right. Note that a height of zero would not cover any
   * pixels.
   */
  get Y2(): number {
    return this._y + this._h;
  }

  /**
   * Get by a named enum
   */
  Get(n: coordName): number {
    switch (n) {
      case coordName.X:
        return this._x;
      case coordName.Y:
        return this._y;
      case coordName.W:
        return this._w;
      case coordName.H:
        return this._h;
      default:
        // this really shouldn't happen, it means we didn't cover all
        // the cases in the switch, which is an enum so...
        throw S5Err.NewError(S5Err.Messages.NotImplemented, `coord {$n}`);
    }
  }

  /**
   * Set by a named enum
   */
  Set(n: coordName, v: number): void {
    switch (n) {
      case coordName.X:
        this._x = v;
        break;
      case coordName.Y:
        this._y = v;
        break;
      case coordName.W:
        this._w = v;
        break;
      case coordName.H:
        this._h = v;
        break;
      default:
        // this really shouldn't happen, it means we didn't cover all
        // the cases in the switch, which is an enum so...
        throw S5Err.NewError(S5Err.Messages.NotImplemented, `coord {$n}`);
    }
  }

  /**
   * Return a new rect that is shifted by a given amount. The value of the
   * width and height are copied from this rect.
   *
   * @param x how much to shift the coordinate system to the right (negative is left)
   * @param y how much to shift the coordinate system to the down (negative is up)
   */
  ToOtherCoordSystem(r: rect): rect {
    return new rect(this._x + r.X, this._y + r.Y, this._w, this._h);
  }

  /**
   * Merge returns the union of two rectangles.  It does not modify either
   * this object or the parameter, it returns the result of the merge.  This
   * function does not sanity check its parameters.
   *
   * @param r other rectangle to be merged with this one.
   */
  Merge(r: rect): rect {
    const x1 = this.X1 < r.X1 ? this.X1 : r.X1;
    const y1 = this.Y1 < r.Y1 ? this.Y1 : r.Y1;
    const x2 = this.X2 > r.X2 ? this.X2 : r.X2;
    const y2 = this.Y2 > r.Y2 ? this.Y2 : r.Y2;

    const h = y2 - y1;
    const w = x2 - x1;
    return new rect(x1, y1, w, h);
  }
}
