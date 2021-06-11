import * as S5Err from '../error';
import * as S5Render from '../render';
import * as S5Conf from '../conf';
import base from './base';
import interactor from './interactor';
import { context2D } from '../render/context';

export default class root extends base implements interactor {
  /**
   * User visible property for having the entire root interactor covered by
   * an image.  If you use this property, the root interactor will not draw
   * until its background image is fully loaded (to prevent FOUC) and will abort if the image
   * cannot be found.  Set to '' if you don't want to use an image.
   */
  _backgroundImage = '';

  /**
   * User visible property to set the id to look for in the DOM.  This element
   * is where the i
   */
  _elementId = 'seven5';

  get BackgroundImage(): string {
    return this._backgroundImage;
  }

  set BackgroundImage(s: string) {
    if (this.backgroundBitmap) {
      this.backgroundBitmap = null;
    } else {
      // we haven't loaded any background yet, so turn off display while we do
      this.el.setAttribute('hidden', 'true');
    }
    this._backgroundImage = s;
  }

  get ElementId(): string {
    return this._elementId;
  }

  set ElementId(s: string) {
    this._elementId = s;
  }

  el: HTMLCanvasElement;

  public get Element(): HTMLCanvasElement {
    return this.el;
  }

  public set Element(value: HTMLCanvasElement) {
    this.el = value;
  }

  backgroundBitmap: ImageBitmap | null = null;

  mainloopInterval: NodeJS.Timeout | null = null;

  backingStore: OffscreenCanvas;

  constructor() {
    super('root');
    const el = document.getElementById(this._elementId);
    if (el === null) {
      throw S5Err.NewError(S5Err.Messages.RootNotFound);
    }
    if (!(el instanceof HTMLCanvasElement)) {
      throw S5Err.NewError(S5Err.Messages.RootNotCanvas);
    }
    if (!el.getContext) {
      throw S5Err.NewError(S5Err.Messages.CanvasHasNoContext);
    }
    this.el = el;
    // this is the most important, yet least publicized, issue with
    // the HTML canvas object...the resolution and size are independent
    // dimensions.  for now, we fix them to be 1:1
    // https://stackoverflow.com/questions/43853119/javascript-wrong-mouse-position-when-drawing-on-canvas
    this.el.style.width = `${this.el.width}px`;
    this.el.style.height = `${this.el.height}px`;

    // this is awful.  see below for why we cannot use this!
    // this.resetBackingSurface(el.width, el.height);

    // this is where we do the actually drawing
    this.backingStore = new OffscreenCanvas(el.width, el.height);
    // set the coords to avoid confusion
    this.X = 0;
    this.Y = 0;
    this.W = el.width;
    this.H = el.height;
    this._transparentBackground = true;
  }

  /**
   * This implements the root of the drawing pattern.  This causes the actual
   * display to be sent to the screen.
   *
   * @param surface
   */
  RenderSelf(ctx: S5Render.Context): void {
    const bsWidth = this.backingStore.width;
    const bsHeight = this.backingStore.height;

    ctx.MoveTo(0, 0); // try to reset?
    if (this._backgroundImage === '') {
      if (this._transparentBackground) {
        // just use a transparent clear
        ctx.ClearRect(0, 0, bsWidth, bsHeight);
      } else {
        S5Conf.Log('root fill');
        ctx.FillStyle = S5Render.DefaultTheme.Background;
        ctx.FillRect(0, 0, bsWidth, bsHeight);
      }
    } else {
      // we are using a background image... but is it loaded yet?
      if (this.backgroundBitmap === null) {
        S5Render.LoadBitmap(this._backgroundImage, bsWidth, bsHeight)
          .then((b: ImageBitmap): void => {
            this.backgroundBitmap = b;
            this.MarkDirty();
          })
          .catch((reason) => {
            S5Conf.Error(`error caught in background load of image: ${reason}`);
            // we really cannot do anything here because we are running in background
            this.StopMainloop();
          });
        return;
      }
      this.el.removeAttribute('hidden');
      // we have the background image now
      ctx.DrawImage(
        this.backgroundBitmap,
        0,
        0,
        bsWidth,
        bsHeight,
        0,
        0,
        bsWidth,
        bsHeight
      );
    }

    // no children yet
    this.MarkClean();
  }

  resetBackingSurface(width: number, height: number): void {
    this.backingStore = new OffscreenCanvas(width, height);
  }

  /**
   * This is the origin point of a redraw cycle.
   */
  redraw(): void {
    // this can throw
    const backingContext = new context2D(this.backingStore);
    this.ContextSizing(backingContext);
    this.Render(backingContext);

    // copy to screen
    const bitmap = this.backingStore.transferToImageBitmap();
    const ctx = this.el.getContext('bitmaprenderer');
    if (!ctx) {
      throw S5Err.NewError(S5Err.Messages.CanvasHasNoContext, 'bitmaprenderer');
    }
    ctx.transferFromImageBitmap(bitmap);
  }

  // passing slowMotion = true slows down the redraws by a factor of
  // twenty and makes them easier to see for debugging
  StartMainloop(slowMotion: boolean): void {
    if (!window) {
      S5Conf.Error('unable to find a window to start mainloop, giving up!');
      return;
    }
    let prev: (ev: PromiseRejectionEvent) => void;
    if (window.onunhandledrejection) {
      prev = window.onunhandledrejection;
    }
    window.onunhandledrejection = (ev: PromiseRejectionEvent) => {
      S5Conf.Error(`stopping mainloop due to background error`);
      this.StopMainloop();
      if (prev) {
        prev(ev);
      }
    };
    this.Mainloop(); // otherwise, we have to wait one interval
    this.mainloopInterval = global.setInterval(
      () => {
        this.Mainloop();
      },
      slowMotion ? 5000 : 250
    );
  }

  // this is a single iteration
  Mainloop(): void {
    S5Conf.Log('mainloop checking dirty...');
    if (this.IsDirty()) {
      S5Conf.Log('... its dirty, so redraw!');
      this.redraw();
      S5Conf.Log(`and after the redraw? ${this.IsDirty()}`);
    }
  }

  StopMainloop(): void {
    if (!this.mainloopInterval) {
      return; // nothing to do
    }
    global.clearInterval(this.mainloopInterval);
    this.mainloopInterval = null;
  }
}
