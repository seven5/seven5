import * as S5Err from '../error';
import color from './color';
import defaultTheme from './defaulttheme';
import baseline from './baseline';

/**
 * This is thin wrapper over the implementations of drawing context provided
 * by Canvas and OffscreenCanvas.
 */
export default interface context {
  Baseline: baseline;
  BeginPath(): void;
  ClearRect(x: number, y: number, w: number, h: number): void;
  Clip(): void;
  ClosePath(): void;
  DrawArc(
    centerX: number,
    centerY: number,
    radius: number,
    startDeg: number,
    stopDeg: number,
    counterClockwise?: boolean
  ): void;
  DrawImage(img: CanvasImageSource, x: number, y: number): void;
  DrawImage(
    img: CanvasImageSource,
    destinationX: number,
    destinationY: number,
    destW: number,
    destH: number
  ): void;
  DrawImage(
    img: CanvasImageSource,
    destinationX: number,
    destinationY: number,
    destW: number,
    destH: number,
    sourceX: number,
    sourceY: number,
    sourceW: number,
    sourceH: number
  ): void;
  Fill(): void;
  FillRect(x: number, y: number, w: number, h: number): void;
  FillStyle: color;
  FillText(s: string, x: number, y: number): void;
  Font: string;
  LineTo(x: number, y: number): void;
  LineWidth: number;
  MeasureText(s: string): TextMetrics;
  MoveTo(x: number, y: number): void;
  Restore(): void;
  Save(): void;
  Stroke(): void;
  StrokeStyle: color;
  StrokeText(s: string, x: number, y: number): void;
  Translate(x: number, y: number): void;
}

export class context2D implements context {
  impl: OffscreenCanvasRenderingContext2D | CanvasRenderingContext2D;

  set Baseline(b: baseline) {
    this.impl.textBaseline = b;
  }

  get Baseline(): baseline {
    const t = this.impl.textBaseline;
    switch (t) {
      case 'alphabetic':
        return baseline.Alphabetic;
      case 'top':
        return baseline.Top;
      case 'middle':
        return baseline.Middle;
      case 'bottom':
        return baseline.Bottom;
      default:
        throw S5Err.NewError(
          S5Err.Messages.CantCreate,
          `Baseline, unsupported baseline value:${t}`
        );
    }
  }

  /**
   * You must supply either a canvas or an offscreenCanvas as parameter here.
   * @param c
   */
  constructor(c: HTMLCanvasElement | OffscreenCanvas) {
    const tmp = c.getContext('2d');
    if (tmp === null) {
      throw S5Err.NewError(S5Err.Messages.CantCreate, 'rendering context');
    }
    this.impl = tmp;
    this.impl.fillStyle = this._fillStyle.value();
    this.impl.strokeStyle = this._strokeStyle.value();
  }

  BeginPath(): void {
    this.impl.beginPath();
  }

  ClearRect(x: number, y: number, w: number, h: number): void {
    this.impl.clearRect(x, y, w, h);
  }

  Clip(): void {
    this.impl.clip();
  }

  ClosePath(): void {
    this.impl.closePath();
  }

  DrawArc(
    centerX: number,
    centerY: number,
    radius: number,
    startDeg: number,
    stopDeg: number,
    counterClockwise?: boolean
  ): void {
    let ccw = false;
    if (counterClockwise === true) {
      ccw = true;
    }
    this.impl.arc(centerX, centerY, radius, startDeg, stopDeg, ccw);
  }

  DrawImage(
    img: CanvasImageSource,
    destinationX: number,
    destinationY: number,
    destW?: number,
    destH?: number,
    sourceX?: number,
    sourceY?: number,
    sourceW?: number,
    sourceH?: number
  ): void {
    if (!destW || !destH) {
      this.impl.drawImage(img, destinationX, destinationY);
      return;
    }
    if (!sourceX || !sourceY || !sourceW || !sourceH) {
      this.impl.drawImage(img, destinationX, destinationY, destW, destH);
      return;
    }

    this.impl.drawImage(
      img,
      destinationX,
      destinationY,
      destW,
      destH,
      sourceX,
      sourceY,
      sourceW,
      sourceH
    );
  }

  Fill(): void {
    this.impl.fill();
  }

  FillRect(x: number, y: number, w: number, h: number): void {
    this.impl.fillRect(x, y, w, h);
  }

  _fillStyle: color = defaultTheme.Foreground;

  get FillStyle(): color {
    return this._fillStyle;
  }

  set FillStyle(f: color) {
    this._fillStyle = f;
    this.impl.fillStyle = f.value();
  }

  FillText(s: string, x: number, y: number): void {
    this.impl.fillText(s, x, y);
  }

  _font = '10px SansSerif';

  get Font(): string {
    return this._font;
  }

  set Font(s: string) {
    this.impl.font = s;
    this._font = s;
  }

  _lineWidth = 1.0;

  get LineWidth(): number {
    return this._lineWidth;
  }

  set LineWidth(w: number) {
    this._lineWidth = w;
  }

  LineTo(x: number, y: number): void {
    this.impl.lineTo(x, y);
  }

  MeasureText(s: string): TextMetrics {
    return this.impl.measureText(s);
  }

  MoveTo(x: number, y: number): void {
    this.impl.moveTo(x, y);
  }

  Restore(): void {
    this.impl.restore();
  }

  Save(): void {
    this.impl.save();
  }

  Stroke(): void {
    this.impl.stroke();
  }

  _strokeStyle: color = defaultTheme.Foreground;

  get StrokeStyle(): color {
    return this._strokeStyle;
  }

  set StrokeStyle(s: color) {
    this._strokeStyle = s;
    this.impl.strokeStyle = s.value();
  }

  StrokeText(s: string, x: number, y: number): void {
    this.impl.strokeText(s, x, y);
  }

  Translate(x: number, y: number): void {
    this.impl.translate(x, y);
  }
}
