<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\ValueObjects;


class ExternalCommandArgsQueryOptions extends QueryOptions {

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $servicedescription;

    /**
     * @var string
     */
    private $nodeName;

    /**
     * @var string
     */
    private $command_name;

    /**
     * @var int
     */
    private $state = 0;

    /**
     * @var string
     */
    private $output = '';

    /**
     * @var string
     */
    private $comment = '';

    /**
     * @var bool
     */
    private $force = true;

    /**
     * @var bool
     */
    private $broadcast = false;

    /**
     * @var int
     */
    private $start = 0;

    private $end = 0;

    /**
     * @var int
     */
    private $duration = 0;

    /**
     * @var bool
     */
    private $sticky = true;

    /**
     * @var string
     */
    private $author_name = 'Anonymous';

    /**
     * @var int
     */
    private $downtime_id = 0;


    /**
     * ExternalCommandArgsQueryOptions constructor.
     * @param $params
     */
    public function __construct($params) {
        $this->start = time();
        $this->duration = 60 * 60 * 15;
        $this->end = $this->start + $this->duration;

        parent::__construct($params);
        $this->params = $params;

        if (isset($this->params['hostname'])) {
            $this->hostname = $this->params['hostname'];
        }

        if (isset($this->params['servicedescription'])) {
            $this->servicedescription = $params['servicedescription'];
        }

        if (isset($this->params['node_name'])) {
            $this->nodeName = $this->params['node_name'];
        }

        if (isset($this->params['command_name'])) {
            $this->command_name = trim($this->params['command_name']);
        }

        if (isset($this->params['state'])) {
            $this->state = $this->params['state'];
        }

        if (isset($this->params['output'])) {
            $this->output = $this->params['output'];
        }

        if (isset($this->params['comment'])) {
            $this->comment = $this->params['comment'];
        }

        if (isset($this->params['force'])) {
            $this->force = true;
            if ($this->params['force'] === 'false' || $this->params['force'] === '0') {
                $this->force = false;
            }
        }

        if (isset($this->params['broadcast'])) {
            $this->broadcast = true;
            if ($this->params['broadcast'] === 'false' || $this->params['broadcast'] === '0') {
                $this->broadcast = false;
            }
        }

        if (isset($this->params['start'])) {
            $this->start = (int)$this->params['start'];
        }

        if (isset($this->params['end'])) {
            $this->end = (int)$this->params['end'];
        }

        if (isset($this->params['sticky'])) {
            $this->sticky = true;
            if ($this->params['sticky'] === 'false' || $this->params['sticky'] === '0') {
                $this->sticky = false;
            }
        }

        if (isset($this->params['author_name'])) {
            $this->author_name = $this->params['author_name'];
        }

        if (isset($this->params['downtime_id'])) {
            $this->downtime_id = (int)$this->params['downtime_id'];
        }
    }

    /**
     * @return string
     */
    public function getHostname() {
        return $this->hostname;
    }

    /**
     * @return string
     */
    public function getServicedescription() {
        return $this->servicedescription;
    }

    /**
     * @return int
     */
    public function getCommandName() {
        return $this->command_name;
    }

    /**
     * @return string
     */
    public function getNodeName() {
        return $this->nodeName;
    }

    /**
     * @return int
     */
    public function getState() {
        return $this->state;
    }

    /**
     * @return string
     */
    public function getOutput() {
        return $this->output;
    }

    /**
     * @return string
     */
    public function getComment() {
        if(empty($this->comment)){
            return 'empty comment';
        }
        return $this->comment;
    }

    /**
     * @return bool
     */
    public function isForce() {
        return $this->force;
    }

    /**
     * @return bool
     */
    public function isBroadcast() {
        return $this->broadcast;
    }

    /**
     * @return int
     */
    public function getStart() {
        return $this->start;
    }

    /**
     * @return int
     */
    public function getEnd() {
        return $this->end;
    }

    /**
     * @return int
     */
    public function getDuration() {
        return $this->end - $this->start;
    }

    /**
     * @return bool
     */
    public function isSticky() {
        return $this->sticky;
    }

    /**
     * @return string
     */
    public function getAuthorName() {
        return $this->author_name;
    }

    /**
     * @return int
     */
    public function getDowntimeId(){
        return $this->downtime_id;
    }

}
