import base from './base';
import * as S5Conf from '../conf';
import * as S5Render from '../render';
import * as S5Err from '../error';
import * as EV from '../evalvite';
import interactor from './interactor';

let { IdCounter } = S5Conf.Vars;

export default class label extends base {
  _text: EV.Attribute<string>;

  get Text(): string {
    return this._text.Get();
  }

  set Text(s: string) {
    this._text.Set(s);
    this.dirtyDimensions = true;
  }

  baselineOffset = 0;

  dirtyDimensions = true;

  _font: S5Render.Font;

  get Font(): S5Render.Font {
    return this._font;
  }

  set Font(f: S5Render.Font) {
    this._font = f;
    this.dirtyDimensions = true;
    this.MarkDirty();
  }

  // should vpadding and hpadding be attributes? seems like overkill
  _vpadding: EV.Attribute<number> = EV.Simple<number>(0, 'label_vpadding');

  get VPadding(): number {
    return this._vpadding.Get();
  }

  set VPadding(p: number) {
    this._vpadding.Set(p);
    this.dirtyDimensions = true;
    this.MarkDirty();
  }

  _hpadding: EV.Attribute<number> = EV.Simple(2, 'label_hpadding');

  get HPadding(): number {
    return this._hpadding.Get();
  }

  set HPadding(p: number) {
    this._hpadding.Set(p);
    this.dirtyDimensions = true;
    this.MarkDirty();
  }

  constructor(s: string, debugName?: string) {
    super(debugName || `[label ${IdCounter}]`);
    if (!debugName) {
      IdCounter += 1;
    }
    this._text = EV.Simple<string>(s, 'label_text');
    this._transparentBackground = false;
    this._font = S5Render.DefaultTheme.FontFamily.Font(
      S5Render.FontStyle.Regular,
      S5Render.FontWeight.Normal,
      12
    );

    // if it's visible, should be eager
    this._vpadding.Eager = true;
    this._hpadding.Eager = true;
    this._text.Eager = true;

    this._hpadding.AddUpdateable(this);
    this._vpadding.AddUpdateable(this);
    this._text.AddUpdateable(this);
  }

  ContextSizing(ctx: S5Render.Context): void {
    if (this.dirtyDimensions) {
      // xxx fixme(iansmith) -- should be hidden
      ctx.Font = `${this.Font.Style} ${this.Font.Weight} ${this.Font.Size}px ${this.Font.Family}`;
      this.updateDimensions(ctx);
    }
  }

  updateDimensions(ctx: S5Render.Context): void {
    const prev = ctx.Baseline;
    ctx.Baseline = S5Render.Baseline.Bottom;
    const myText = ctx.MeasureText(this._text.Get());
    ctx.Baseline = prev;
    const textWidth = Math.ceil(myText.width);
    this.Coords.AttrW.Set(textWidth + 2 * this._hpadding.Get());
    this.Coords.AttrH.Set(
      myText.fontBoundingBoxAscent +
        myText.fontBoundingBoxDescent +
        2 * this._vpadding.Get()
    );

    this.dirtyDimensions = false;
  }

  Render(ctx: S5Render.Context): void {
    super.RenderNoChildren(ctx);
  }

  FillBackground(ctx: S5Render.Context): void {
    ctx.FillStyle = S5Render.DefaultTheme.Background;
    // ctx.FillStyle = S5Render.DefaultTheme.Background;
    ctx.FillRect(0, 0, this.W, this.H);
  }

  // without the ContextSizing, this RenderSelf could have the effect
  // of recomputing width and height of this interactor if dirtyDimensions
  // is true. The ContextSizing pass is used to remedy this problem.
  RenderSelf(ctx: S5Render.Context): void {
    if (this.dirtyDimensions) {
      S5Conf.Warn(
        `RenderSelf() of label ${this.DebugName()} has unsized content!`
      );
      this.updateDimensions(ctx);
    }
    if (!this._transparentBackground) {
      this.FillBackground(ctx);
    }
    this.RenderText(ctx);
    this.dirty = false;
  }

  RenderText(ctx: S5Render.Context): void {
    const prev = ctx.Font;
    // xxx fixme(iansmith) -- should be hidden
    ctx.Font = `${this.Font.Style} ${this.Font.Weight} ${this.Font.Size}px ${this.Font.Family}`;
    // ctx.StrokeStyle = new RGBColor(0, 0, 0);
    ctx.FillStyle = S5Render.DefaultTheme.Foreground;
    ctx.FillText(
      this._text.Get(),
      this._hpadding.Get(),
      this._vpadding.Get() + S5Render.FontHeight(this.Font)
    );
    ctx.Font = prev;
  }

  // eslint-disable-next-line class-methods-use-this,@typescript-eslint/no-unused-vars
  AddChildAt(i: interactor, x: number, y: number): void {
    throw S5Err.NewError(S5Err.Messages.NoKids);
  }

  // eslint-disable-next-line class-methods-use-this,@typescript-eslint/no-unused-vars
  AddChild(i: interactor): void {
    throw S5Err.NewError(S5Err.Messages.NoKids);
  }

  // eslint-disable-next-line class-methods-use-this,
  set Children(k: interactor[]) {
    throw S5Err.NewError(S5Err.Messages.NoKids);
  }
}
