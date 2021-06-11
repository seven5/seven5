import * as S5Event from '../event';
import interactor from './interactor';
import * as S5Render from '../render';
import * as EV from '../evalvite';
import * as S5Conf from '../conf';
import * as S5Err from '../error';
import attrRect from './attrrect';

let { IdCounter } = S5Conf.Vars;

export default class base implements interactor {
  /**
   * Coords holds the x,y,w,h for this interactor.  Note that the x and y
   * are in the *parent's* coordinate system because in our coordinate system
   * they would also be 0,0 and pointless.  This is the storage for the coords
   * but the getters and setters are part of interactor.
   */
  Coords: attrRect;

  /// //////////////
  // Coordinates
  /// /////////////
  get X(): number {
    return this.Coords.Get(S5Render.CoordName.X);
  }

  set X(n: number) {
    this.Coords.Set(S5Render.CoordName.X, n);
  }

  get Y(): number {
    return this.Coords.Get(S5Render.CoordName.Y);
  }

  set Y(n: number) {
    this.Coords.Set(S5Render.CoordName.Y, n);
  }

  get W(): number {
    return this.Coords.Get(S5Render.CoordName.W);
  }

  set W(n: number) {
    this.Coords.Set(S5Render.CoordName.W, n);
  }

  get H(): number {
    return this.Coords.Get(S5Render.CoordName.H);
  }

  set H(n: number) {
    this.Coords.Set(S5Render.CoordName.H, n);
  }

  RectToParentCoordSystem(r: S5Render.Rect): S5Render.Rect {
    return new S5Render.Rect(r.X + this.X, r.Y + this.Y, r.Width, r.Height);
  }

  /// /////////////
  // Dirty handling, damage rectangles
  /// ////////////

  /* this controls dirtiness ONLY, you can be dirty and have no
  damage rectangles if a CHILD of yours has damage. */
  dirty = true;

  /**
   * dirtyArea holds the rectangle that needs a repaint with respect to *this*
   * interactor only.
   */
  damageArea: S5Render.Rect | null = null;

  /**
   * MarkDirty() is a function that updates a given interactor's dirty state
   * _and_ may do a multitude of other things.  It is always wise to use
   * MarkDirty() to indicate a redraw is needed -- and possibly to use
   * AddDamage() if you want to indicate what needs to be redrawn.
   *
   * MarkDirty() propagates to the parent of this interactor if there is
   * a parent of this interactor.
   */
  MarkDirty(): void {
    if (this.dirty) {
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `S5DEBUG:${this.DebugName()}: no need to mark dirty, already dirty`
        );
      }
      return;
    }
    this.dirty = true;
    if (this.Parent === null) {
      // making this an error might detect some programming errors but
      // it will certainly cause false positives when people are doing
      // the setup of their structures before they are on screen.
      if (S5Conf.Vars.Debug) {
        S5Conf.Log(
          `S5DEBUG:${this.DebugName()}: marked dirty, no parent to mark`
        );
      }
      return;
    }
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(
        `S5DEBUG:${this.DebugName()}: marked dirty, marking parent dirty...`
      );
    }
    this.Parent.MarkDirty();
  }

  /**
   * This should be called when *this* interactor needs a subsection
   * repainted.  You can just make it the bounds of the interactor
   * if you just want the whole thing redrawn.  This is used to
   * choose clipping rectangles to make redraw faster.
   *
   * Some interactors may want to propagate damage to their parents
   * and some parents may want to interrogate damage of their children.
   */
  AddDamage(r: S5Render.Rect): void {
    if (!this.damageArea) {
      this.damageArea = r;
    } else {
      this.damageArea = this.damageArea.Merge(r);
    }
  }

  /**
   * This returns the currently damaged area or falsey (null) if no damage.
   */
  Damage(): S5Render.Rect | null {
    return this.damageArea;
  }

  /**
   * This clears all the damage for this interactor. You should use this
   * after you finish RenderSelf().
   */
  ClearDamage(): void {
    this.damageArea = null;
  }

  MarkClean(): void {
    this.dirty = false;
  }

  IsDirty(): boolean {
    return this.dirty;
  }

  /// /////////////
  // Debug name
  /// ////////////
  dname: string;

  DebugName(): string {
    return this.dname;
  }

  /// /////////////
  // Parents and children
  /// ////////////

  _parent: interactor | null = null;

  get Parent(): interactor | null {
    return this._parent;
  }

  set Parent(i: interactor | null) {
    this._parent = i;
  }

  _children = [] as interactor[];

  get Children(): interactor[] {
    return this._children;
  }

  set Children(c: interactor[]) {
    this._children = c;
  }

  HasChildren(): boolean {
    return !!this._children && this._children.length > 0;
  }

  AddChildAt(i: interactor, x: number, y: number): void {
    if (i.Parent) {
      throw S5Err.NewError(S5Err.Messages.ParentExists, 'AddChildAt');
    }
    this._children.push(i);
    i.Parent = this;
    i.X = x;
    i.Y = y;
    i.MarkDirty();
    this.MarkDirty();
  }

  /**
   * Currently, this is a no-op but it probably should be doing work
   * to figure out which attributes might still be connected to
   * this interactor and to other parts of the attribute graph and
   * disconnect them. This is called _after_ the interactor has
   * been removed from the child list.
   *
   * @param i the interactor to disconnect
   */
  // eslint-disable-next-line
  DisconnectChild(i: interactor): void {}

  /**
   * Remove given child from child list.  Returns false if the
   * child cannot be found.
   * @param i interactor to remove
   */
  RemoveChild(i: interactor): boolean {
    const targetIndex = this.Children.indexOf(i);
    if (targetIndex < 0) {
      return false;
    }
    const dead = this.Children[targetIndex];
    this.Children.splice(targetIndex, 1);
    dead.Parent = null;
    this.DisconnectChild(dead);
    this.MarkDirty();
    return true;
  }

  /**
   * ClearChildren gets rid of all children of an interactor.
   */
  ClearChildren(): void {
    // need to call RemoveChild because often it has logic that
    // "undoes" what was done on add child.
    while (this.Children.length > 0) {
      this.RemoveChild(this.Children[0]);
    }
    this.MarkDirty();
  }

  AddChild(i: interactor): void {
    this.Children.push(i);
    i.Parent = this;
    i.MarkDirty();
    this.MarkDirty();
  }

  /**
   * MyIndex returns the index of this interactor in its parent.
   * This throws if either there is no parent or the parent's child
   * list does not contain this.
   */
  MyIndex(): number {
    return base.ChildIndex(this);
  }

  /**
   * ChildIndex returns the index of this provided interactor in its parent.
   * This throws if either there is no parent or the parent's child
   * list does not contain the interactor given.  If you provide the
   * second parameter as true, it will return a negative value if the
   * child cannot be found.  We could use 'indexOf()' but we try to
   * detect some programming errors while doing this work.
   */
  static ChildIndex(i: interactor, allowMissing?: boolean): number {
    if (i.Parent === null) {
      throw S5Err.NewError(S5Err.Messages.NoParent, 'ChildIndex');
    }
    let found = false;
    let ci = 0;
    i.Parent.Children.forEach((c, index) => {
      if (c === i) {
        if (found) {
          throw S5Err.NewError(
            S5Err.Messages.BadState,
            "ChildIndex() found child in parent's child list twice"
          );
        }
        found = true;
        ci = index;
      }
    });
    if (!found) {
      if (allowMissing) {
        return -1;
      }
      throw S5Err.NewError(
        S5Err.Messages.BadState,
        "MyIndex() could not locate child in parent's child list"
      );
    }
    return ci;
  }

  /// /////////////
  // Rendering
  /// ////////////
  _transparentBackground = false;

  get TransparentBackground(): boolean {
    return this._transparentBackground;
  }

  set TransparentBackground(b: boolean) {
    this._transparentBackground = b;
  }

  ContextSizing(ctx: S5Render.Context): void {
    this.Children.forEach((c) => c.ContextSizing(ctx));
  }

  /**
   * Default behavior of this implementation is to draw the parent content, if
   * any, behind the children.  Override this method to do different/more complex
   * drawing of children.
   *
   * @param surface
   */
  Render(ctx: S5Render.Context): void {
    this.SetClippingRect(ctx);
    this.RenderSelf(ctx);
    this._children.forEach((child: interactor) => {
      ctx.Save();
      ctx.Translate(child.X, child.Y);
      child.Render(ctx);
      ctx.Restore();
    });
    this.ClearDamage();
    this.MarkClean();
  }

  /**
   * Tricky: This method is designed to be used as a replacement
   * for the Render() method that defaults to descending into children.
   * If your interactor cannot have children, its best to use this
   * method as the implementation of Render() via super.RenderNoChildren().
   */
  RenderNoChildren(ctx: S5Render.Context): void {
    this.SetClippingRect(ctx);
    this.RenderSelf(ctx);
  }

  /**
   * This is the method new interactors should use to draw themselves.
   */
  RenderSelf(ctx: S5Render.Context): void {
    if (!this._transparentBackground) {
      ctx.FillStyle = S5Render.DefaultTheme.Background;
      ctx.FillRect(0, 0, this.W, this.H);
    }
    // we use width and height as coords here, because we want lines that cross
    // the whole area of the interactor
    ctx.StrokeStyle = S5Render.DefaultTheme.Foreground;
    ctx.BeginPath();
    ctx.MoveTo(0, 0);
    ctx.LineTo(this.W, this.H);
    ctx.MoveTo(this.W, 0);
    ctx.LineTo(0, this.H);
    ctx.Stroke();
  }

  /**
   * This defaults to setting the clipping rect to the area of thi
   * interactor defined by X,Y and W,H.
   */
  SetClippingRect(ctx: S5Render.Context): void {
    let w = this.W;
    let h = this.H;
    let x = 0;
    let y = 0;
    const dmg = this.Damage();
    if (dmg !== null) {
      w = dmg.Width;
      h = dmg.Height;
      x = dmg.X;
      y = dmg.Y;
    }
    ctx.BeginPath();
    ctx.MoveTo(x, y);
    ctx.LineTo(w, y);
    ctx.LineTo(w, h);
    ctx.LineTo(x, h);
    ctx.LineTo(x, y);

    // const fs = ctx.FillStyle;
    // const t = new S5Render.RGBColor(0, 0, 0);
    // t.Alpha = 0.0;
    // ctx.FillStyle = t;
    // ctx.Fill();
    ctx.Clip();
    // ctx.FillStyle = fs;
  }

  /**
   * Create an instance of the defaultInteractor.
   * @param useAttributeCoords set to true if you prefer x,y,w,h to
   * use a simple attribute rather than a variable.
   */
  constructor(debugName?: string) {
    this.dname = debugName || `[default interactor ${IdCounter}`;
    if (!debugName) {
      IdCounter += 1;
    }
    /*
     * The default values of x, y, width, and height have been chosen carefully. They
     * are designed to make it easy to reveal bugs when combined with the default
     * implementation of renderSelf() in defaultInteractor.
     */
    this.Coords = new attrRect();
    this.Coords.AttrX = EV.Simple(20, `${this.DebugName()}.X`);
    this.Coords.AttrY = EV.Simple(10, `${this.DebugName()}.Y`);
    this.Coords.AttrW = EV.Simple(15, `${this.DebugName()}.W`);
    this.Coords.AttrH = EV.Simple(30, `${this.DebugName()}.H`);
    EV.Bind(this.Coords.AttrX, this);
    EV.Bind(this.Coords.AttrY, this);
    EV.Bind(this.Coords.AttrW, this);
    EV.Bind(this.Coords.AttrH, this);
    this.MarkDirty();
  }

  /**
   * Called when a number attribute is hooked to this interactor.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any,class-methods-use-this,@typescript-eslint/no-unused-vars,@typescript-eslint/no-empty-function
  ConnectAttribute(a: EV.Attribute<any>): void {}

  /**
   * Called when a number attribute is unhooked from this interactor.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any,class-methods-use-this,@typescript-eslint/no-unused-vars,@typescript-eslint/no-empty-function
  DisconnectAttribute(a: EV.Attribute<any>): void {}

  /**
   * Update value from an attribute. Right now we just always mark the
   * object as dirty with the assumption we can redraw our way out of this.
   * Protocol is: both values are given for a normal change of the attribute,
   * like a Set().  Only first value given if we had no previous value, meaning
   * this is the first Set() of the value.  If neither value is provided
   * then this is a notification that the value of the attribute a is
   * dirty and, if needed for the correct visuals, should be requested.
   */
  // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types,@typescript-eslint/no-explicit-any,@typescript-eslint/no-unused-vars
  UpdateValue(a: EV.Attribute<any>, v?: any, old?: any): void {
    if (S5Conf.Vars.Debug) {
      S5Conf.Log(
        `S5DEBUG: ${a.DebugName()}:updated value on interactor: ${a.DebugName()} to ${v}`
      );
    }
    this.MarkDirty();
  }

  /**
   * Default implementation of "inside" is bounding rectangle. Coords
   * are in the interactor's coord system.
   */
  Inside(x: number, y: number): boolean {
    if (x >= 0 && x < this.W) {
      if (y >= 0 && y < this.H) {
        return true;
      }
    }
    return false;
  }

  /**
   * Picks is a recursive (downward, pre-order) call to create a PickList.  The
   * pl parameter is being "built up" during the recursive pass.  The list
   * is ordered back to front.  The event's coords are mutated as the
   * the pick list goes around the tree.
   */
  Picks(e: S5Event.Event, pl: S5Event.PickList): void {
    const x = e.X;
    const y = e.Y;

    if (this.Inside(x, y)) {
      pl.AppendHit(this);
    }
    if (this.HasChildren()) {
      this.Children.forEach((c: S5Event.Pickable) => {
        e.Translate(c.X, c.Y); // adjust for child
        c.Picks(e, pl);
        e.Translate(-c.X, -c.Y); // reset
      });
    }
  }
}
