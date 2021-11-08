import React from 'react';
import 'bootstrap/dist/css/bootstrap.min.css';
import Button from 'react-bootstrap/Button';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Image from 'react-bootstrap/Image';
import clock from './clock.png';
import stocksstocks from './stock.png';
import pga from './pga.png';
import sys from './server.png';
import Form from 'react-bootstrap/Form';
import { MatrixPostRet, JSONToStatus, JumpToBoard } from './util';
import * as pb from './basicboard/basicboard_pb';

class BasicBoard extends React.Component {
    constructor(props) {
        super(props);
        var path = this.props.name;
        if (this.props.path) {
            path = this.props.path;
        }
        this.state = {
            "path": path,
            "status": new pb.Status(),
        };
    }
    async componentDidMount() {
        await this.getStatus();
    }
    getStatus = async () => {
        console.log("BasicBoard GetStatus", this.state.path + "/board.v1.BasicBoard/GetStatus");
        await MatrixPostRet(this.state.path + "/board.v1.BasicBoard/GetStatus", '{}').then((resp) => {
            if (resp.ok) {
                return resp.text();
            }
            throw resp;
        }).then((data) => {
            var dat = JSONToStatus(data);
            this.setState({
                "status": dat,
            });
        });
    }

    updateStatus = async () => {
        var req = new pb.SetStatusReq();
        req.setStatus(this.state.status);
        await MatrixPostRet(this.state.path + "/board.v1.BasicBoard/SetStatus", JSON.stringify(req.toObject()));
        await this.getStatus();
    }

    logosrc() {
        if (this.props.name == "clock") {
            return clock;
        } else if (this.props.name == "stocks") {
            return stocksstocks;
        } else if (this.props.name === "pga") {
            return pga;
        } else if (this.props.name === "sys") {
            return sys;
        }
    }
    render() {
        return (
            <Container fluid>
                <Row className="text-center"><Col><Image src={this.logosrc()} style={{ height: '100px', width: 'auto' }} onClick={() => { JumpToBoard(this.props.name); this.props.doSync(); }} fluid /></Col></Row>
                <Row className="text-left">
                    <Col>
                        <Form.Switch id={this.props.name + "enabler"} label="Enable/Disable" checked={this.state.status.getEnabled()}
                            onChange={() => { this.state.status.setEnabled(!this.state.status.getEnabled()); this.updateStatus(); }} />
                    </Col>
                </Row>
                <Row className="text-left">
                    <Col>
                        <Form.Switch id={this.props.name + "scroll"} label="Scroll Mode" checked={this.state.status.getScrollEnabled()}
                            onChange={() => { this.state.status.setScrollEnabled(!this.state.status.getScrollEnabled()); this.updateStatus(); }} />
                    </Col>
                </Row>
                <Row className="text-left">
                    <Col>
                        <Button variant="primary" onClick={() => { JumpToBoard(this.props.name); this.updateStatus(); this.props.doSync(); }}>Jump</Button>

                    </Col>
                </Row>
            </Container>
        )
    }
}

export default BasicBoard;