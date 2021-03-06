/**
 * Copyright 2018 Shift Devices AG
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { Component, h, RenderableProps } from 'preact';
import { CopyableInput } from '../../../components/copy/Copy';
import { Button } from '../../../components/forms';
import { QRCode } from '../../../components/qrcode/qrcode';
import { translate, TranslateProps } from '../../../decorators/translate';
import { apiGet, apiPost } from '../../../utils/request';

interface ProvidedProps {
    info: SigningConfigurationInterface;
    code: string;
}

export interface SigningConfigurationInterface {
    scriptType: 'p2pkh' | 'p2wpkh-p2sh' | 'p2pkh';
    keypath: string;
    threshold: number;
    xpubs: string[];
    address: string;
}

interface State {
    canVerifyExtendedPublicKey: number[]; // holds a list of keystores which support secure verification
}

type Props = ProvidedProps & TranslateProps;

class SigningConfiguration extends Component<Props, State> {
    constructor(props) {
        super(props);
        this.state = ({ canVerifyExtendedPublicKey: [] });
        this.canVerifyExtendedPublicKeys();
    }

    private canVerifyExtendedPublicKeys = () => {
        apiGet(`account/${this.props.code}/can-verify-extended-public-key`).then(canVerifyExtendedPublicKey => {
            this.setState({ canVerifyExtendedPublicKey });
        });
    }

    private verifyExtendedPublicKey = (index: number) => {
        apiPost(`account/${this.props.code}/verify-extended-public-key`, index);
    }

    public render({ t, info }: RenderableProps<Props>, { canVerifyExtendedPublicKey }: State) {
        return (
        // TODO: add info if single or multisig, and threshold.
        <div>
            { info.address ?
                <div>
                    <strong>
                        {t('accountInfo.address')}
                    </strong><br />
                    <QRCode data={info.address} />
                    <CopyableInput value={info.address} />
                </div>
                    :
                info.xpubs.map((xpub, index) => {
                    return (
                        <div key={xpub}>
                            <strong>
                                {t('accountInfo.extendedPublicKey')}
                                {info.xpubs.length > 1 && (' #' + (index + 1))}
                            </strong><br />
                            <QRCode data={xpub} />
                            <CopyableInput value={xpub} />
                            { canVerifyExtendedPublicKey.includes(index) ?
                                <Button primary onClick={() => this.verifyExtendedPublicKey(index)}>
                                    {t('accountInfo.verify')}
                                </Button> : '' }
                        </div>
                    );
                })
            }
        </div>
    ); }
}

const HOC = translate<ProvidedProps>()(SigningConfiguration);
export { HOC as SigningConfiguration };
