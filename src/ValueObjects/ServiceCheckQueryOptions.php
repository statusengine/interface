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


class ServiceCheckQueryOptions extends QueryOptions {

    private $columnsForOrder = [
        'hostname',
        'service_description',
        'state',
        'is_hardstate',
        'output',
        'start_time',
        'latency'
    ];

    /**
     * @var string
     */
    private $order = 'start_time';

    /**
     * @var string
     */
    private $direction = 'desc';

    /**
     * @var array
     */
    private $state = [];

    /**
     * @var array
     */
    private $stateMatch = [
        'ok' => 0,
        'warning' => 1,
        'critical' => 2,
        'unknown' => 3
    ];
    /**
     * @var string
     */
    private $output__like = '';

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $service_description;

    /**
     * ServiceCheckQueryOptions constructor.
     * @param $params
     * @throws \Exception
     */
    public function __construct($params) {
        parent::__construct($params);

        if (isset($params['order'])) {
            if (!in_array($params['order'], $this->columnsForOrder)) {
                throw new \Exception('Invalid column for order by condition!');
            }
            $this->order = $params['order'];
        }

        if (isset($params['direction'])) {
            //asc is class default
            if ($params['direction'] == 'desc') {
                $this->direction = 'desc';
            }
        }

        if (isset($this->params['state']) && is_array($this->params['state'])) {
            $_state = [];
            foreach ($this->params['state'] as $state) {
                if (isset($this->stateMatch[$state])) {
                    $_state[] = $this->stateMatch[$state];
                }
            }
            $this->state = $_state;
        }

        if (isset($this->params['output__like'])) {
            $this->output__like = $this->params['output__like'];
        }

        if (isset($this->params['hostname'])) {
            $this->hostname = $params['hostname'];
        }

        if (isset($this->params['service_description'])) {
            $this->service_description = $params['service_description'];
        }

    }

    /**
     * @return string
     */
    public function getOrder() {
        return $this->order;
    }

    /**
     * @return string
     */
    public function getDirection() {
        return $this->direction;
    }

    public function getStateFilter() {
        return $this->state;
    }

    public function sizeOfStateFilter() {
        return sizeof($this->state);
    }

    /**
     * @return string
     */
    public function getOutputLike() {
        return $this->output__like;
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
    public function getServiceDescription() {
        return $this->service_description;
    }

}
