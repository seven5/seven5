export enum messages {
  RootNotFound = 'root element not found in DOM',
  RootNotCanvas = 'root element in DOM is not a canvas',
  CanvasHasNoContext = 'root element is a canvas, but will not return context for drawing',
  BadDrawingInput = 'value supplied was not valid for drawing',
  NotImplemented = 'method not implemented',
  NetworkError = 'network error',
  CantCreate = 'unable to create API object',
  NoParent = 'interactor has no parent for operation',
  NoFunc = 'evalvite computed attribute has no computation function assigned',
  Constrained = "interactor's value is computed and cannot be set",
  NullAttr = 'cannot compute value from a null attribute (you can use a simple attribute if you want a constant)',
  NoKids = 'interactor does not allow children',
  ParentExists = 'interactor already has a parent',
  MismatchedParameters = 'number of input attributes does not match number of parameters to computation',
  DuplicateOutgoing = 'attempting to add an outgoing edge that already exists',
  NotAllowed = 'attribute operation not allowed',
  BadState = 'interactor or attribute in bad state',
  BadIndex = 'index of array attribute out of range',
  NoPolicy = 'unable to find input policy',
  NoAgent = 'unable to find input dispatch agent',
  CoordMissing = 'provided only X and not Y or vice versa',
}

export function newError(e: messages, detail?: string): Error {
  if (detail) {
    return new Error(`${e}: ${detail}`);
  }
  return new Error(e);
}
