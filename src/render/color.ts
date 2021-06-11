import * as S5Err from '../error';

export default interface color {
  Alpha: number;
  /**
   * Returns the string version of the color, formatted for use in web apps.
   */
  value(): string;
}

export class rgbColor implements color {
  static White = new rgbColor(255, 255, 255);

  static Black = new rgbColor(0, 0, 0);

  r: number;

  g: number;

  b: number;

  constructor(r: number, g: number, b: number) {
    if (r < 0 || r > 255) {
      throw S5Err.NewError(S5Err.Messages.BadDrawingInput, `red: ${r}`);
    }
    if (g < 0 || g > 255) {
      throw S5Err.NewError(S5Err.Messages.BadDrawingInput, `green:${g}`);
    }
    if (b < 0 || b > 255) {
      throw S5Err.NewError(S5Err.Messages.BadDrawingInput, `blue:${b}`);
    }
    this.r = Math.floor(r);
    this.g = Math.floor(g);
    this.b = Math.floor(b);
  }

  _alpha: number | undefined = undefined;

  get Alpha(): number {
    if (this._alpha === undefined) {
      return 1.0;
    }
    return this._alpha;
  }

  set Alpha(n: number) {
    if (n < 0.0 || n > 1.0) {
      throw S5Err.NewError(S5Err.Messages.BadDrawingInput, `alpha:${n}`);
    }
    this._alpha = n;
  }

  /**
   * Returns the rgb value of this, possibly with alpha value, so this can be used
   * in web-related calls.  Note the string returned is carefully formatted to
   * to be web-legal.
   */
  value(): string {
    if (this._alpha === undefined) {
      return `rgb(${this.r},${this.g},${this.b})`;
    }
    return `rgb(${this.r},${this.g},${this.b},${this._alpha})`;
  }
}
