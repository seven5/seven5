import label from './label';
import * as S5Render from '../render';
import * as S5Event from '../event';
import * as EV from '../evalvite';
import * as S5Conf from '../conf';

import color from '../render/color';

export default class button extends label {
  mouseIsDown: EV.Attribute<boolean> = EV.Simple<boolean>(false);

  fill: EV.Attribute<color> = EV.Computed<color>(
    (b: boolean) => {
      if (b) {
        return S5Render.DefaultTheme.Foreground;
      }
      return S5Render.DefaultTheme.Background;
    },
    [this.mouseIsDown],
    'fillColor'
  );

  stroke: EV.Attribute<color> = EV.Computed<color>(
    (b: boolean) => {
      if (b) {
        return S5Render.DefaultTheme.Background;
      }
      return S5Render.DefaultTheme.Foreground;
    },
    [this.mouseIsDown],
    'strokeColor'
  );

  FillBackgroundOrStrokeBoundary(ctx: S5Render.Context, isFill: boolean): void {
    const sum = ((this.H - this.VPadding) / 2) ** 2 + this.HPadding ** 2;
    const radius = Math.round(Math.sqrt(sum));

    ctx.FillStyle = this.fill.Get();
    ctx.StrokeStyle = this.stroke.Get();

    // top
    ctx.BeginPath();
    ctx.MoveTo(radius, 0);
    // ctx.BeginPath();
    ctx.LineTo(this.W - 1 - radius, 0);

    // right
    ctx.DrawArc(
      this.W - 1 - radius,
      this.H / 2 - 1,
      radius,
      button.degToRad(270.0),
      button.degToRad(90.0)
    );

    // bottom
    ctx.LineTo(radius, this.H - 1);

    // left
    ctx.DrawArc(
      radius,
      this.H / 2 - 1,
      radius,
      button.degToRad(90.0),
      button.degToRad(270.0)
    );

    if (isFill) {
      ctx.Fill();
    } else {
      ctx.Stroke();
    }
  }

  RenderSelf(ctx: S5Render.Context): void {
    S5Conf.Log('Render self of button');
    this.FillBackgroundOrStrokeBoundary(ctx, !this.TransparentBackground);
    this.RenderText(ctx);
    this.dirty = false;
  }

  static degToRad(d: number): number {
    const deg = d % 360;
    return (deg / 360.0) * 2 * Math.PI;
  }

  constructor(text: string, debugName?: string) {
    super(text, debugName);

    // set some good looking defaults
    this._hpadding.Set(15);
    this._vpadding.Set(10);

    // if it's visible, should be eager
    this.fill.Eager = true;
    this.stroke.Eager = true;

    // connect attributes to this interactor
    this.fill.AddUpdateable(this);
    this.stroke.AddUpdateable(this);

    S5Conf.Log('fill is ', this.fill);
    S5Conf.Log('stroke is ', this.stroke);
  }

  //
  // Interaction Protocol for Clickable
  //

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ClickStart(e: S5Event.Event): void {
    S5Conf.Log('ClickStart');
    S5Conf.Vars.Debug = true;
    this.mouseIsDown.Set(true);
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  ClickEnd(e: S5Event.Event): void {
    S5Conf.Log('ClickEnd');
    this.mouseIsDown.Set(false);
  }
}
