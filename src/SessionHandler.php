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

namespace Statusengine;


class SessionHandler {

    public function __construct() {
        if(session_status() == PHP_SESSION_NONE){
            session_start();
        }
    }

    /**
     * @param string $key
     * @return bool
     */
    public function has($key = ''){
        return isset($_SESSION[$key]);
    }

    /**
     * @param string $key
     * @param mixed|null $value
     */
    public function set($key = '', $value = null){
        $_SESSION[$key] = $value;
    }

    /**
     * @param string $key
     * @return mixed|false
     */
    public function get($key = ''){
        if($this->has($key)){
            return $_SESSION[$key];
        }
        return false;
    }

    /**
     * @return string
     */
    public function getSessionId(){
        return session_id();
    }

    public function start(){
        session_start();
    }

    public function writeClose(){
        session_write_close();
    }

    public function destroy(){
        session_destroy();
    }

    /**
     * @return string
     */
    public function getUsername(){
        if($this->has('username')){
            return $this->get('username');
        }
        return 'Anonymous';
    }

}