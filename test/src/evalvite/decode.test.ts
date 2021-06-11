import * as EV from '../../../src/evalvite';
import {expect} from 'chai';

describe('decode', ()=> {
  it('base cases', ()=> {
    expect(EV.DecodeValue(Object.keys({})).length).to.be.equal(0);
    expect(EV.DecodeValue([])).to.be.an('array').that.is.empty;

    expect(EV.DecodeValue(7)).to.be.equal(7);
    expect(EV.DecodeValue('fleazil')).to.be.equal('fleazil');
    expect(EV.DecodeValue({foo:'bar'})).to.deep.equal({foo:'bar'});
    const s = EV.Simple<number>(-12);
    expect(EV.DecodeValue(s)).to.be.equal(-12);
    expect(EV.DecodeValue({s:s})).to.deep.equal({s:-12});
    expect(EV.DecodeValue({x:{y:{z:s}}})).to.deep.equal({x:{y:{z:-12}}});
    const f = EV.Simple<string>('frobnitz');
    expect(EV.DecodeValue(f)).to.be.equal('frobnitz');
    expect(EV.DecodeValue({s:f})).to.deep.equal({s:'frobnitz'});
    const a = [ 1, 2, 3, 4];
    expect(EV.DecodeValue(a)).to.deep.equal([1,2,3,4]);
    const b = [ 1, 2, 3, 4, s];
    expect(EV.DecodeValue(b)).to.deep.equal([1,2,3,4,-12]);
    expect(EV.DecodeValue({'a':a,n:{f:f, s:s}})).to.deep.equal({a:[1,2,3,4],n:{f:'frobnitz',s:-12}});
  });
});
