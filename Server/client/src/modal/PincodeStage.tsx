import * as React from 'react';
import './EmailLoginModal.css';
import SrcReactCodeInput from 'react-code-input';

interface Props {
  email: string;
  token: string;
  onCompletion: (success: boolean, token?: string) => void;
}

interface State {
  errorMessage?: string;
  code?: string;
}

class PincodeStage extends React.Component<Props> {
  state: State;

  constructor(props: Props) {
    super(props);
    this.state = {};
  }

  isCodeValid() {
    let re = /^\d+$/;
    return this.state.code !== undefined && this.state.code.length === 9 && re.test(this.state.code);
  }

  onValidateCode() {
    let self = this;
    const token = this.props.token;
    const uri = 'https://api.iagd.evilsoft.net/auth';

    const code = this.state.code as string;
    fetch(uri, {
        method: 'POST',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'multipart/form-data'
        },
        body: `key=${token}&code=${code}`
      }
    )
      .then((response) => {
        if (!response.ok) {
          console.log(response);
          throw Error(`Got response ${response.status}, ${response.statusText}`);
        }
        return response;
      })
      .then((response) => response.json())
      .then((json) => {
        if (json.token !== undefined) {
          this.props.onCompletion(json.type, json.token);
        }
        else {
          console.warn('Attempted to authenticate code, but the result status was undefined.');
          this.props.onCompletion(false);
        }
      })
      .catch((error) => {
        console.warn(error);
        self.setState({errorMessage: `${error}`});
      });
  }


  render() {
    let re = /^\d+$/;
    const showNonNumericError = this.state.code !== undefined && !re.test(this.state.code);

    return (
      <div>
        <h2>An E-Mail has been sent to <span className="email-label">{this.props.email}</span> with the verification
          code</h2>
        <div className="code-input">
          <SrcReactCodeInput
            type="text"
            inputMode={'numeric'}
            fields={9}
            name='codeinput'
            onChange={(e) => this.setState({code: e})}
          />
          {showNonNumericError && <div className="alert alert-warning">The code can only consist of numbers</div>}
          <input
            className={!this.isCodeValid() ? 'form-control btn btn-default' : 'form-control btn btn-primary'}
            type="button"
            value="Verify"
            disabled={!this.isCodeValid()}
            onClick={() => this.onValidateCode()}
          />
        </div>
      </div>
    );
  }

}

export default PincodeStage;

