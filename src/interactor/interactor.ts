// we would like to do this:
// import * as S5Render from '../render';
// xxx fixme(iansmith)
// but doing that creates an import cycle, so we import only the things
// we need from S5Render
import rect from '../render/rect';
import context from '../render/context';

// normal imports
import * as EV from '../evalvite';
import * as S5Render from '../render';
import * as S5Event from '../event';

export default interface interactor {
  X: number;
  Y: number;
  W: number;
  H: number;
  /**
   * dirtyArea holds the rectangle that needs a repaint with respect to *this*
   * interactor only.  If an interactor is dirty and has no damage area,
   * the whole interactor needs a redraw.
   */
  damageArea: rect | null;

  /**
   * _parent is a pointer to the parent interactor. May be null, especially
   * for the root of the tree.
   */
  Parent: interactor | null;

  /**
   * Given x,y in this interactor's coord system, compute a point in the parent's
   * coordinate system that references the same pixel.  The rect returned will have
   * same width and height and the values supplied.  This function does not modify
   * either this interactor's coords or the parameter's.
   * @param r Rect to transform
   */
  RectToParentCoordSystem(r: rect): rect;

  /**
   * MarkDirty() is a function that updates a given interactor's dirty state
   * _and_ may do a multitude of other things.
   */
  MarkDirty(): void;

  /**
   * This should be called to indicate that you are aware of a subregion
   * of this interactor that needs a redraw. If the interactor is
   * dirty *and* has no damage, then the entire interactor needs a redraw.
   */
  AddDamage(r: S5Render.Rect): void;

  /**
   * Damage() returns the currently damaged region of this interactor
   * or null if there no subregion to be redrawn.
   */
  Damage(): S5Render.Rect | null;

  /**
   * MarkClean is the opposite of MarkDirty. Usually, this called
   * at the end of render().
   *
   */
  MarkClean(): void;

  /**
   * ContextSizing is used when doing "inside out" layouts.  These are
   * layouts where the size of children dictate the size of parents.  This
   * is problematic when you have any interactor that cannot compute
   * its correct size without known the exact context it is going to
   * be using to draw.  The primary example of this are interactors
   * that use fonts, so things like text width and text height cannot
   * be known without a context to help compute these values.  This should
   * _only_ used to compute size information, it should not do drawing.
   */
  ContextSizing(ctx: S5Render.Context): void;

  /**
   * This renders the visuals for this interactor and all of its children.
   * In the default implementation, RenderSelf() is called before drawing children
   * by calling their Render().
   * @param surface the drawing context
   */
  Render(ctx: context): void;

  /**
   * This renders the visuals for *this* interactor to the given surface.  In
   * the default implementation, Render() will call RenderSelf().
   * @param ctx the drawing context
   */
  RenderSelf(ctx: context): void;

  /**
   * This is used to set the clipping rect on the current interactor
   * to the interactor's bounds *or* the region indicated as damaged.
   * Override with care, as this allows you to set the clipping to
   * "somebody else's" screen space.
   *
   */
  SetClippingRect(ctx: S5Render.Context): void;

  /**
   * Array of children of this interactor. Note that this can be
   * empty and will be for leaf nodes.
   */
  Children: interactor[];

  /**
   * Test if we have children.
   */
  HasChildren():boolean;
  /**
   * AddChildAt adds an element to the end of the child array for this
   * interactor.  It is located, relative to the parent's origin at (x,y).
   * Note that you should use AddChild() if you have attributes placed in
   * the X or Y slots.
   * @param i
   * @param x
   * @param y
   */
  AddChildAt(i: interactor, x: number, y: number): void;

  /**
   * AddChild adds a child to the child list of parent, but doesn't
   * set any of it's coordinates.  This is mostly useful when you have
   * computed attributes on the interactors coords.
   */
  AddChild(i: interactor): void;
  /**
   * Removes a child given a pointer to it.  Returns false if the
   * child could not be found (using === on the child list).
   * @param i
   */
  RemoveChild(i: interactor): boolean;
  /**
   * ClearChildren removes all the children of an interactor.
   */
  ClearChildren(): void;
  /**
   * MyIndex returns the index of this interactor in its parent.
   * This throws if either there is no parent or the parent's child
   * list does not contain this.
   */
  MyIndex(): number;

  /**
   * Picks is a recursive (downward) call to create a PickList.
   */
  Picks(e: S5Event.Event, pl: S5Event.PickList): void;

  /**
   * Called when a number attribute is hooked to this interactor.
   */
  // eslint-disable-next-line  @typescript-eslint/no-explicit-any
  ConnectAttribute(a: EV.Attribute<any>): void;
  /**
   * Called when a number attribute is unhooked from this interactor.
   */
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  DisconnectAttribute(a: EV.Attribute<any>): void;
  /**
   * Update value from an attribute.   Note that if you want to have
   * a better type for the value here, you'll need to do bookkeeping
   * that maps a particular attribute to a particular _type_ of value
   * and then call some other method with a more specific type of value.
   */
  // eslint-disable-next-line  @typescript-eslint/no-explicit-any
  UpdateValue(a: EV.Attribute<any>, v: any): void;
}
